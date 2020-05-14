package statemgr

import (
	"github.com/iotaledger/wasp/packages/committee"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/state"
	"time"
)

// respond to sync request 'GetStateUpdate'
func (sm *stateManager) EventGetStateUpdateMsg(msg *committee.GetStateUpdateMsg) {
	if stateUpd, err := state.LoadStateUpdate(sm.committee.ScId(), msg.StateIndex); err == nil {
		_ = sm.committee.SendMsg(msg.SenderIndex, committee.MsgStateUpdate, hashing.MustBytes(&committee.StateUpdateMsg{
			StateUpdate: stateUpd,
		}))
	}
	// no need for action because it doesn't change of the state
}

// respond to state update msg.
// It collects state updates while waiting for the anchoring state transaction
// only are stored updates to the current solid variable state
func (sm *stateManager) EventStateUpdateMsg(msg *committee.StateUpdateMsg) {
	if !sm.addPendingStateUpdate(msg.StateUpdate) {
		return
	}
	if msg.StateUpdate.StateTransactionId() != sctransaction.NilId && !msg.FromVM {
		// state update contains state transaction in it
		// so we need to ask for the corresponding state transaction
		// except when it is calculated locally by the VM
		sm.asyncRequestForStateTransaction(msg.StateUpdate.StateTransactionId(), sm.committee.ScId(), msg.StateUpdate.StateIndex())
		sm.syncMessageDeadline = time.Now().Add(parameters.SyncPeriodBetweenSyncMessages)
	}
	sm.takeAction()
}

// triggered whenever new state transaction arrives
func (sm *stateManager) EventStateTransactionMsg(msg committee.StateTransactionMsg) {
	stateBlock, ok := msg.Transaction.State()
	if !ok {
		return
	}
	sm.CheckSynchronizationStatus(stateBlock.StateIndex())

	if sm.solidVariableState == nil || stateBlock.StateIndex() != sm.solidVariableState.StateIndex()+1 {
		// only interested for the state transaction to verify latest state update
		return
	}
	sm.nextStateTransaction = msg.Transaction

	sm.takeAction()
}

func (sm *stateManager) EventTimerMsg(msg committee.TimerTick) {
	if msg%10 == 0 {
		sm.takeAction()
	}
}
