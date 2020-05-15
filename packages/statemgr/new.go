// statemgr package implements object which is responsible for the smart contract
// ledger state to be synchronized and validated
package statemgr

import (
	"github.com/iotaledger/wasp/packages/committee"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/state"
	"time"
)

type stateManager struct {
	committee committee.Committee

	// pending state updates are candidates to confirmation by the state transaction
	// which leads to the state transition
	// the map key is hash of the variable state which is a result of applying the state update
	// to the solid variable state
	pendingStateUpdates map[hashing.HashValue]*pendingStateUpdate

	// state transaction with +1 state index from the state index of solid variable state
	// it may be nil if does not exist or not fetched yet
	nextStateTransaction *sctransaction.Transaction

	// last variable state stored in the database
	// it may be nil at bootstrap when origin variable state is calculated
	solidVariableState state.VariableState

	// largest state index seen from other messages. If this index is more than 1 step ahead then
	// the solid one, state is not synced
	largestEvidencedStateIndex uint32

	// synchronization status. It is reset when state becomes synchronized

	// pseudo-random permutation of peer indices. Serves a sequence in which peers are queried for state updates
	permutationOfPeers []uint16
	// next peer permutationOfPeers[permutationIndex] is a next peer will be asked for ths state uodate
	permutationIndex uint16
	// the timeout deadline for sync inquiries
	syncMessageDeadline time.Time
}

type pendingStateUpdate struct {
	// state update, not validated yet
	stateUpdate state.StateUpdate
	// resulting variable state applied to the solidVariableState
	nextVariableState state.VariableState
}

func New(committee committee.Committee) committee.StateManager {
	ret := &stateManager{
		committee:           committee,
		pendingStateUpdates: make(map[hashing.HashValue]*pendingStateUpdate),
	}
	go ret.initLoadState()

	return ret
}

// initial loading of the solid state
func (sm *stateManager) initLoadState() {
	var err error

	scid := sm.committee.ScId()
	// load last variable state from the database
	sm.solidVariableState, err = state.LoadVariableState(scid)
	if err != nil {
		log.Errorf("can't load variable state for scid %s: %v", scid.String(), err)

		return
	}
	stateIndex := uint32(0)
	if sm.solidVariableState != nil {
		stateIndex = sm.solidVariableState.StateIndex()
	}
	// if sm.solidVariableState == nil it may be an origin state

	// load solid state update from db with the state index taken from the variable state
	// state index is 0 if variable state doesn't exist in the DB
	stateUpdate, err := state.LoadStateUpdate(scid, stateIndex)
	if err != nil {
		log.Errorf("dismiss committee: can't load state update index %d for scid %s=%v", stateIndex, scid.String(), err)

		sm.committee.Dismiss()
		return
	}
	if !sm.addPendingStateUpdate(stateUpdate) {
		panic("assertion failed: sm.addPendingStateUpdate(stateUpdate)")
	}

	// open msg queue for the committee
	sm.committee.OpenQueue()

	// request last state transaction to update sync status
	go sm.findLastStateTransaction(sm.committee.ScId())

	// async load state transaction for the current state update
	sm.asyncRequestForStateTransaction(stateUpdate.StateTransactionId(), sm.committee.ScId(), stateIndex)

}
