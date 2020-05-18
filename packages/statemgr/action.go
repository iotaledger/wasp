package statemgr

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/committee"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util"
	"time"
)

func (sm *stateManager) takeAction() {
	if sm.checkStateTransition() {
		return
	}
	sm.requestStateUpdateFromPeerIfNeeded()
}

func (sm *stateManager) checkStateTransition() bool {
	if sm.nextStateTransaction == nil {
		return false
	}
	// among pending state updates we locate the one, consistent with the next state transaction
	varStateHash := sm.nextStateTransaction.MustState().VariableStateHash()
	pending, ok := sm.pendingBatches[*varStateHash]
	if !ok {
		// corresponding batch wasn't found among pending state updates
		// state transition is not possible
		return false
	}
	// found corresponding pending batch
	// it is approved by the nextStateTransaction, because hashes of coincide
	// batch is marked with approving stat tx id. It doesn't change essence bytes of the batch
	if !pending.batch.IsCommitted() {
		pending.batch.Commit(sm.nextStateTransaction.ID())
	} else {
		// should be same transaction hash
		if !(sm.nextStateTransaction.ID() == pending.batch.StateTransactionId()) {
			sm.log.Panicf("assertion failed: sm.nextStateTransaction.ID() == pending.batch.StateTransactionId()")
		}
	}

	// save the new state and mark requests as processed
	if err := pending.nextVariableState.Commit(sm.committee.Address(), pending.batch); err != nil {
		sm.log.Errorf("failed to save next state #%d", pending.batch.StateIndex())
		return false
	}

	prevStateIndex := ""
	if sm.solidVariableState.StateIndex() > 0 {
		prevStateIndex = fmt.Sprintf("#%d", sm.solidVariableState.StateIndex()-1)
	}
	sm.log.Infof("state transition #%s --> #%d sc addr %s",
		prevStateIndex, sm.solidVariableState.StateIndex(), sm.committee.Address().String())

	saveTx := sm.nextStateTransaction

	// update state manager variables to the new state
	sm.solidVariableState = pending.nextVariableState
	sm.nextStateTransaction = nil
	sm.pendingBatches = make(map[hashing.HashValue]*pendingBatch) // clean pending batches
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

const syncPeriodBetweenSyncMessages = 1 * time.Second

func (sm *stateManager) requestStateUpdateFromPeerIfNeeded() {
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
	data := hashing.MustBytes(&committee.GetBatchMsg{
		PeerMsgHeader: committee.PeerMsgHeader{
			StateIndex: sm.solidVariableState.StateIndex() + 1,
		},
	})
	// send messages until first without error
	for i := uint16(0); i < sm.committee.Size(); i++ {
		targetPeerIndex := sm.permutationOfPeers[sm.permutationIndex]
		if err := sm.committee.SendMsg(targetPeerIndex, committee.MsgGetBatch, data); err == nil {
			break
		}
		sm.permutationIndex = (sm.permutationIndex + 1) % sm.committee.Size()
		sm.syncMessageDeadline = time.Now().Add(syncPeriodBetweenSyncMessages)
	}
}

// index of evidenced state index is passed to record the largest one.
// This is needed to check synchronization status. If some state index is more than
// 1 behind the largest, node is not synced
// function returns if the message with idx must be passed to consensus operator, which works only with
// state indices of current or next state
func (sm *stateManager) CheckSynchronizationStatus(idx uint32) bool {
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

func (sm *stateManager) isSynchronized() bool {
	return sm.largestEvidencedStateIndex-sm.solidVariableState.StateIndex() <= 1
}

// adding batch of state updates to the 'pending' map
func (sm *stateManager) addPendingBatch(batch state.Batch) bool {
	var varState state.VariableState
	var err error

	if sm.solidVariableState != nil {
		if batch.StateIndex() != sm.solidVariableState.StateIndex()+1 {
			return false
		}
	} else {
		if batch.StateIndex() != 0 {
			return false
		}
	}

	varState = state.NewVariableState(sm.solidVariableState)

	err = varState.Apply(batch)
	if err != nil {
		sm.log.Warn(err)
		return false
	}
	sm.pendingBatches[*varState.Hash()] = &pendingBatch{
		batch:             batch,
		nextVariableState: varState,
	}
	return true
}
