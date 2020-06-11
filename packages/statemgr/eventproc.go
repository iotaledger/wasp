package statemgr

import (
	"github.com/iotaledger/wasp/packages/committee"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/plugins/publisher"
)

// respond to sync request 'GetStateUpdate'
func (sm *stateManager) EventGetBatchMsg(msg *committee.GetBatchMsg) {
	//sm.log.Debugw("EventGetBatchMsg",
	//	"sender index", msg.SenderIndex,
	//	"state index", msg.StateIndex,
	//)
	addr := sm.committee.Address()
	batch, err := state.LoadBatch(addr, msg.StateIndex)
	if err != nil || batch == nil {
		//sm.log.Debugw("EventGetBatchMsg: can't find batch",
		//	"sender index", msg.SenderIndex,
		//	"state index", msg.StateIndex,
		//	"addr", addr.String(),
		//)
		// can't load batch, can't respond
		return
	}
	sm.log.Debugw("EventGetBatchMsg: sending batch",
		"target peer", msg.SenderIndex,
		"state index", msg.StateIndex,
		"state tx", batch.StateTransactionId().String(),
		"size", batch.Size(),
		"essence", batch.EssenceHash().String(),
		"ts", batch.Timestamp(),
	)
	err = sm.committee.SendMsg(msg.SenderIndex, committee.MsgBatchHeader, util.MustBytes(&committee.BatchHeaderMsg{
		PeerMsgHeader: committee.PeerMsgHeader{
			StateIndex: msg.StateIndex,
		},
		Size:               batch.Size(),
		StateTransactionId: batch.StateTransactionId(),
		Timestamp:          batch.Timestamp(),
	}))
	if err != nil {
		return
	}
	batch.ForEach(func(stateUpdate state.StateUpdate) bool {
		err = sm.committee.SendMsg(msg.SenderIndex, committee.MsgStateUpdate, util.MustBytes(&committee.StateUpdateMsg{
			PeerMsgHeader: committee.PeerMsgHeader{
				StateIndex: msg.StateIndex,
			},
			StateUpdate: stateUpdate,
		}))
		sh := util.GetHashValue(stateUpdate)
		sm.log.Debugw("EventGetBatchMsg: sending stateUpdate", "hash", sh.String())
		return err == nil
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
		len(sm.syncedBatch.stateUpdates) == int(msg.Size) &&
		sm.syncedBatch.ts == msg.Timestamp {
		return // no need to start from scratch
	}
	sm.syncedBatch = &syncedBatch{
		stateIndex:   msg.StateIndex,
		stateUpdates: make([]state.StateUpdate, msg.Size),
		stateTxId:    msg.StateTransactionId,
		ts:           msg.Timestamp,
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
		sm.log.Errorf("bad state update message")
		return
	}
	sh := util.GetHashValue(msg.StateUpdate)
	sm.log.Debugw("EventStateUpdateMsg: receiving stateUpdate", "hash", sh.String())

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
	batch.WithStateIndex(sm.syncedBatch.stateIndex).
		WithTimestamp(sm.syncedBatch.ts).
		WithStateTransaction(sm.syncedBatch.stateTxId)

	sm.log.Debugw("EventStateUpdateMsg: reconstructed batch",
		"essence", batch.EssenceHash().String(),
		"ts", batch.Timestamp(),
	)
	sm.syncedBatch = nil
	go func() {
		sm.committee.ReceiveMessage(committee.PendingBatchMsg{
			Batch: batch,
		})
	}()
	sm.takeAction()
}

// triggered whenever new state transaction arrives
func (sm *stateManager) EventStateTransactionMsg(msg committee.StateTransactionMsg) {
	stateBlock, ok := msg.Transaction.State()
	if !ok {
		return
	}

	vh := stateBlock.VariableStateHash()
	sm.log.Debugw("EventStateTransactionMsg",
		"txid", msg.ID().String(),
		"state index", stateBlock.StateIndex(),
		"state hash", vh.String(),
	)

	sm.CheckSynchronizationStatus(stateBlock.StateIndex())

	if sm.solidStateValid {
		if stateBlock.StateIndex() != sm.solidVariableState.StateIndex()+1 {
			sm.log.Debugf("only interested for the state transaction to verify latest state update")
			return
		}
	} else {
		if sm.solidVariableState == nil {
			if stateBlock.StateIndex() != 0 {
				sm.log.Debugf("sm.solidVariableState == nil && stateBlock.StateIndex() != 0")
				return
			}
		} else {
			if stateBlock.StateIndex() != sm.solidVariableState.StateIndex() {
				sm.log.Debugf("sm.solidVariableState == nil && stateBlock.StateIndex() != sm.solidVariableState.StateIndex()")
				return
			}
		}
	}
	sm.nextStateTransaction = msg.Transaction

	// new state transaction accepted by the state manager
	publisher.Publish("statetx", "in",
		sm.committee.Address().String(),
		sm.nextStateTransaction.ID().String(),
	)

	sm.takeAction()
}

func (sm *stateManager) EventPendingBatchMsg(msg committee.PendingBatchMsg) {
	sm.log.Debugw("EventPendingBatchMsg",
		"state index", msg.Batch.StateIndex(),
		"size", msg.Batch.Size(),
		"txid", msg.Batch.StateTransactionId().String(),
		"essence", msg.Batch.EssenceHash().String(),
		"ts", msg.Batch.Timestamp(),
	)

	sm.addPendingBatch(msg.Batch)
	sm.takeAction()
}

func (sm *stateManager) EventTimerMsg(msg committee.TimerTick) {
	if msg%10 == 0 {
		sm.takeAction()
	}
}
