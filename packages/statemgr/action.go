package statemgr

import (
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/packages/committee"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util"
	"time"
)

func (sm *StateManager) takeAction() {
	if sm.checkStateTransition() {
		return
	}
	sm.requestStateUpdateFromPeerIfNeeded()
}

func (sm *StateManager) checkStateTransition() bool {
	if sm.nextStateTransaction == nil {
		return false
	}
	// among pending state updates we locate the one, consistent with the next state transaction
	varStateHash := sm.nextStateTransaction.MustState().VariableStateHash()
	pending, ok := sm.pendingStateUpdates[varStateHash]
	if !ok {
		// corresponding state update wasn't found among pending state updates
		return false
	}
	// found corresponding pending state
	// it is approved by the nextStateTransaction
	pending.stateUpdate.SetStateTransactionId(sm.nextStateTransaction.ID())

	// save the new state and mark requests as processed
	reqIds := sm.nextStateTransaction.MustState().RequestIds()
	if err := state.SaveStateToDb(pending.stateUpdate, pending.nextVariableState, reqIds); err != nil {
		log.Errorf("failed to save state #%d: %v", pending.stateUpdate.StateIndex(), err)
		return false
	}

	prevStateIndex := ""
	if sm.solidVariableState.StateIndex() > 0 {
		prevStateIndex = fmt.Sprintf("#%d", sm.solidVariableState.StateIndex()-1)
	}
	log.Infof("state transition %s --> #%d scid %s", prevStateIndex, sm.solidVariableState.StateIndex())

	saveTx := sm.nextStateTransaction

	// update state manager variables to the new state
	sm.solidVariableState = pending.nextVariableState
	sm.nextStateTransaction = nil
	sm.pendingStateUpdates = make(map[hashing.HashValue]*pendingStateUpdate) // clean pending state updates
	sm.permutationOfPeers = util.GetPermutation(sm.committee.Size(), varStateHash.Bytes())
	sm.permutationIndex = 0
	sm.syncMessageDeadline = time.Now() // if not synced then immediately

	// if synchronized, notify consensus operator about state transition
	if sm.isSynchronized() {
		sm.committee.ReceiveMessage(&committee.StateTransitionMsg{
			VariableState:    sm.solidVariableState,
			StateTransaction: saveTx,
		})
	}
	return true
}

func (sm *StateManager) requestStateUpdateFromPeerIfNeeded() {
	if sm.solidVariableState == nil {
		return
	}
	if sm.isSynchronized() {
		// state is synced, no need for more info
		return
	}
	// not synced
	if !sm.syncMessageDeadline.Before(time.Now()) {
		// not time yet for the next message
		return
	}
	// it is time to ask for the next state update to next peer in the permutation
	sm.permutationIndex = (sm.permutationIndex + 1) % sm.committee.Size()
	data := hashing.MustBytes(&committee.GetStateUpdateMsg{
		PeerMsgHeader: committee.PeerMsgHeader{
			StateIndex: sm.solidVariableState.StateIndex() + 1,
		},
	})
	// send messages until first without error
	for i := uint16(0); i < sm.committee.Size(); i++ {
		targetPeerIndex := sm.permutationOfPeers[sm.permutationIndex]
		if err := sm.committee.SendMsg(targetPeerIndex, committee.MsgGetStateUpdate, data); err == nil {
			break
		}
		sm.permutationIndex = (sm.permutationIndex + 1) % sm.committee.Size()
		sm.syncMessageDeadline = time.Now().Add(parameters.SyncPeriodBetweenSyncMessages)
	}
}

// index of evidenced state index is passed to record the largest one.
// This is needed to check synchronization status. If some state index is more than
// 1 behind the largest, node is not synced
// function returns if the message with idx must be passed to consensus operator, which works only with
// state indices of current or next state
func (sm *StateManager) CheckSynchronizationStatus(idx uint32) bool {
	// synced state is when current state index is behind
	// the largestEvidencedStateIndex no more than by 1 point
	wasSynchronized := sm.isSynchronized()
	if idx > sm.largestEvidencedStateIndex {
		sm.largestEvidencedStateIndex = idx
	}
	if !sm.isSynchronized() && wasSynchronized {
		sm.syncMessageDeadline = time.Now()
	}
	currentStateIndex := uint32(0)
	if sm.solidVariableState != nil {
		currentStateIndex = sm.solidVariableState.StateIndex()
	}
	return idx == currentStateIndex || idx == currentStateIndex+1
}

func (sm *StateManager) isSynchronized() bool {
	return sm.largestEvidencedStateIndex-sm.solidVariableState.StateIndex() <= 1
}

// async loads state transaction from DB and validates it
// posts 'StateTransactionMsg' to the committee upon success
func (sm *StateManager) asyncRequestForStateTransaction(txid transaction.ID, scid sctransaction.ScId, stateIndex uint32) {
	go func() {
		tx, err := sctransaction.LoadTx(txid)
		if err != nil {
			log.Errorf("can't load state tx",
				"txid", txid.String(),
				"stateIndex", stateIndex,
				"scid", scid.String(),
			)
			return
		}
		stateBlock, ok := tx.State()
		if !ok {
			log.Errorf("not a state tx",
				"txid", txid.String(),
				"stateIndex", stateIndex,
				"scid", scid.String(),
			)
			return
		}
		if *stateBlock.ScId() != scid || stateBlock.StateIndex() != stateIndex {
			log.Errorf("unexpected state tx data",
				"txid", txid.String(),
				"stateIndex", stateIndex,
				"scid", scid.String(),
			)
			return
		}
		// posting to the committee's queue
		sm.committee.ReceiveMessage(committee.StateTransactionMsg{
			Transaction: tx,
		})
	}()
}

func (sm *StateManager) findLastStateTransaction(scid sctransaction.ScId) {
	// finds transaction, which owns output with colored toke scid.Color()
	// notifies committee about it
	// posting to the committee's queue
	sm.committee.ReceiveMessage(committee.StateTransactionMsg{
		Transaction: nil, // TODO stub
	})
}

// adding state update to the 'pending' map
func (sm *StateManager) addPendingStateUpdate(stateUpdate state.StateUpdate) bool {
	var varState state.VariableState
	if sm.solidVariableState != nil {
		if stateUpdate.StateIndex() != sm.solidVariableState.StateIndex()+1 {
			// only interested in updates to the current state
			return false
		}
		varState = sm.solidVariableState.Apply(stateUpdate)
	} else {
		if stateUpdate.StateIndex() != 0 {
			// in the origin, only interested in updates with index 0
			return false
		}
		varState = state.CreateOriginVariableState(stateUpdate)
	}

	stateHash := hashing.GetHashValue(varState)
	existingRecord, ok := sm.pendingStateUpdates[stateHash]
	if ok && existingRecord.stateUpdate.StateTransactionId() != sctransaction.NilId {
		// corresponding pending update already exist
		return false
	}
	sm.pendingStateUpdates[stateHash] = &pendingStateUpdate{
		stateUpdate:       stateUpdate,
		nextVariableState: varState,
	}
	return true
}
