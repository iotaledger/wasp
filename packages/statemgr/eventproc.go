package statemgr

import (
	"github.com/iotaledger/wasp/packages/committee"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/state"
	"time"
)

// respond to sync request 'GetStateUpdate'
func (sm *stateManager) EventGetBatchMsg(msg *committee.GetBatchMsg) {
	batch, err := state.LoadBatch(sm.committee.Address(), msg.StateIndex)
	if err != nil {
		// can't load batch, can't respond
		return
	}
	err = sm.committee.SendMsg(msg.SenderIndex, committee.MsgBatchHeader, hashing.MustBytes(&committee.BatchHeaderMsg{
		PeerMsgHeader: committee.PeerMsgHeader{
			StateIndex: msg.StateIndex,
		},
		Size:               batch.Size(),
		StateTransactionId: batch.StateTransactionId(),
	}))
	if err != nil {
		return
	}
	batch.ForEach(func(stateUpdate state.StateUpdate) bool {
		err = sm.committee.SendMsg(msg.SenderIndex, committee.MsgStateUpdate, hashing.MustBytes(&committee.StateUpdateMsg{
			PeerMsgHeader: committee.PeerMsgHeader{
				StateIndex: msg.StateIndex,
			},
			StateUpdate: stateUpdate,
		}))
		return err == nil
	})
}

func (sm *stateManager) EventBatchHeaderMsg(msg *committee.BatchHeaderMsg) {
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

// respond to state update msg.
// It collects state updates while waiting for the anchoring state transaction
// only are stored updates to the current solid variable state
func (sm *stateManager) EventStateUpdateMsg(msg *committee.StateUpdateMsg) {
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
		WithTimestamp(time.Unix(0, sm.syncedBatch.ts)).
		WithStateTransaction(sm.syncedBatch.stateTxId)

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
	sm.log.Debugw("received transaction")

	stateBlock, ok := msg.Transaction.State()
	if !ok {
		return
	}
	sm.CheckSynchronizationStatus(stateBlock.StateIndex())

	vh := stateBlock.VariableStateHash()
	sm.log.Debugw("received state transaction",
		"txid", msg.ID().String(),
		"state index", stateBlock.StateIndex(),
		"state hash", vh.String(),
	)

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
	sm.takeAction()
}

func (sm *stateManager) EventPendingBatchMsg(msg committee.PendingBatchMsg) {
	sm.log.Debugf("EventPendingBatchMsg")

	sm.addPendingBatch(msg.Batch)
	sm.takeAction()
}

func (sm *stateManager) EventTimerMsg(msg committee.TimerTick) {
	if msg%10 == 0 {
		sm.takeAction()
	}
}
