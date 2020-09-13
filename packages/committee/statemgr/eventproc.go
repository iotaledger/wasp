package statemgr

import (
	"github.com/iotaledger/wasp/packages/committee"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util"
)

// EventPingPongMsg reacts to the PinPong message
func (sm *stateManager) EventPingPongMsg(msg *committee.PingPongMsg) {
	sm.pingPongReceived(msg.SenderIndex)
	if msg.RSVP && sm.solidStateValid {
		// only respond if current solid state is validated
		sm.respondPongToPeer(msg.SenderIndex)
	}
	sm.log.Debugw("EventPingPongMsg",
		"sender", msg.SenderIndex,
		"state", msg.StateIndex,
		"rsvp", msg.RSVP,
		"numPongs", sm.numPongs(),
	)
}

// EventGetBatchMsg is a request for a batch while syncing
func (sm *stateManager) EventGetBatchMsg(msg *committee.GetBatchMsg) {
	sm.log.Debugw("EventGetBatchMsg",
		"sender index", msg.SenderIndex,
		"state index", msg.StateIndex,
	)
	addr := sm.committee.Address()
	batch, err := state.LoadBatch(addr, msg.StateIndex)
	if err != nil || batch == nil {
		// can't load batch, can't respond
		return
	}

	sm.log.Debugf("EventGetBatchMsg for state index #%d --> peer %d", msg.StateIndex, msg.SenderIndex)

	err = sm.committee.SendMsg(msg.SenderIndex, committee.MsgBatchHeader, util.MustBytes(&committee.BatchHeaderMsg{
		PeerMsgHeader: committee.PeerMsgHeader{
			StateIndex: msg.StateIndex,
		},
		Size:               batch.Size(),
		StateTransactionId: batch.StateTransactionId(),
	}))
	if err != nil {
		return
	}
	batch.ForEach(func(batchIndex uint16, stateUpdate state.StateUpdate) bool {
		err = sm.committee.SendMsg(msg.SenderIndex, committee.MsgStateUpdate, util.MustBytes(&committee.StateUpdateMsg{
			PeerMsgHeader: committee.PeerMsgHeader{
				StateIndex: msg.StateIndex,
			},
			StateUpdate: stateUpdate,
			BatchIndex:  batchIndex,
		}))
		sh := util.GetHashValue(stateUpdate)
		sm.log.Debugw("EventGetBatchMsg: sending stateUpdate", "hash", sh.String())
		return true
	})
}

func (sm *stateManager) EventBatchHeaderMsg(msg *committee.BatchHeaderMsg) {
	sm.log.Debugw("EventBatchHeaderMsg",
		"sender", msg.SenderIndex,
		"state index", msg.StateIndex,
		"size", msg.Size,
		"state tx", msg.StateTransactionId.String(),
	)
	if sm.syncedBatch != nil &&
		sm.syncedBatch.stateIndex == msg.StateIndex &&
		sm.syncedBatch.stateTxId == msg.StateTransactionId &&
		len(sm.syncedBatch.stateUpdates) == int(msg.Size) {
		return // no need to start from scratch
	}
	sm.syncedBatch = &syncedBatch{
		stateIndex:   msg.StateIndex,
		stateUpdates: make([]state.StateUpdate, msg.Size),
		stateTxId:    msg.StateTransactionId,
	}
}

// response to the state update msg.
// It collects state updates while waiting for the anchoring state transaction
func (sm *stateManager) EventStateUpdateMsg(msg *committee.StateUpdateMsg) {
	sm.log.Debugw("EventStateUpdateMsg",
		"sender", msg.SenderIndex,
		"state index", msg.StateIndex,
		"batch index", msg.BatchIndex,
	)
	if sm.syncedBatch == nil {
		return
	}
	if sm.syncedBatch.stateIndex != msg.StateIndex {
		return
	}
	if int(msg.BatchIndex) >= len(sm.syncedBatch.stateUpdates) {
		sm.log.Errorf("bad batch index in the state update message")
		return
	}
	sh := util.GetHashValue(msg.StateUpdate)
	sm.log.Debugf("EventStateUpdateMsg: receiving stateUpdate batch index: %d hash: %s",
		msg.BatchIndex, sh.String())

	sm.syncedBatch.stateUpdates[msg.BatchIndex] = msg.StateUpdate
	sm.syncedBatch.msgCounter++

	if int(sm.syncedBatch.msgCounter) < len(sm.syncedBatch.stateUpdates) {
		// some are missing
		return
	}
	// check if whole batch already received
	for _, su := range sm.syncedBatch.stateUpdates {
		if su == nil {
			// some state updates are missing
			return
		}
	}
	// the whole batch received
	batch, err := state.NewBatch(sm.syncedBatch.stateUpdates)
	if err != nil {
		sm.log.Errorf("failed to create batch: %v", err)
		sm.syncedBatch = nil
		return
	}
	batch.WithStateIndex(sm.syncedBatch.stateIndex).WithStateTransaction(sm.syncedBatch.stateTxId)

	sm.log.Debugf("EventStateUpdateMsg: reconstructed batch %s", batch.String())

	sm.syncedBatch = nil
	go sm.committee.ReceiveMessage(committee.PendingBatchMsg{
		Batch: batch,
	})
	sm.takeAction()
}

// EventStateTransactionMsg triggered whenever new state transaction arrives
// the state transaction may be confirmed or not
func (sm *stateManager) EventStateTransactionMsg(msg *committee.StateTransactionMsg) {
	stateBlock, ok := msg.Transaction.State()
	if !ok {
		// should not happen: must have state block
		return
	}

	vh := stateBlock.StateHash()
	sm.log.Debugw("EventStateTransactionMsg",
		"txid", msg.ID().String(),
		"state index", stateBlock.StateIndex(),
		"state hash", vh.String(),
	)

	sm.EvidenceStateIndex(stateBlock.StateIndex())

	if sm.solidStateValid {
		if stateBlock.StateIndex() != sm.solidState.StateIndex()+1 {
			sm.log.Debugf("skip state transaction: expected with state index #%d, got #%d, Txid: %s",
				sm.solidState.StateIndex()+1, stateBlock.StateIndex(), msg.ID().String())
			return
		}
	} else {
		if sm.solidState == nil {
			// pre-origin
			if stateBlock.StateIndex() != 0 {
				sm.log.Debugf("sm.solidState == nil && stateBlock.StateIndex() != 0")
				return
			}
		} else {
			if stateBlock.StateIndex() != sm.solidState.StateIndex() {
				sm.log.Debugf("sm.solidState == nil && stateBlock.StateIndex() != sm.solidState.StateIndex()")
				return
			}
		}
	}
	sm.nextStateTransaction = msg.Transaction

	sm.takeAction()
}

func (sm *stateManager) EventPendingBatchMsg(msg committee.PendingBatchMsg) {
	sm.log.Debugw("EventPendingBatchMsg",
		"state index", msg.Batch.StateIndex(),
		"size", msg.Batch.Size(),
		"txid", msg.Batch.StateTransactionId().String(),
		"batch essence", msg.Batch.EssenceHash().String(),
		"ts", msg.Batch.Timestamp(),
	)

	sm.addPendingBatch(msg.Batch)
	sm.takeAction()
}

func (sm *stateManager) EventTimerMsg(msg committee.TimerTick) {
	if msg%2 == 0 {
		sm.takeAction()
	}
}
