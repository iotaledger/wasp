package statemgr

import (
	"github.com/iotaledger/wasp/packages/committee"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/plugins/nodeconn"
	"time"
)

func (sm *stateManager) takeAction() {
	if sm.checkStateApproval() {
		return
	}
	sm.requestStateTransactionIfNeeded()
	sm.requestStateUpdateFromPeerIfNeeded()
}

// checks the state of the state manager. If one of pending state update batches is confirmed
// by the nextStateTransaction changes the state to the next
func (sm *stateManager) checkStateApproval() bool {
	if sm.nextStateTransaction == nil {
		return false
	}
	// among pending state update batches we locate the one which
	// is approved by the transaction
	varStateHash := sm.nextStateTransaction.MustState().VariableStateHash()
	lst, ok := sm.pendingBatches[varStateHash]
	if !ok {
		// corresponding batch wasn't found among pending state updates
		// transaction doesn't approve anything
		return false
	}
	// there are pending batches with the var state hash which it approves
	// find pending batch who has the same state tx id, i.e. is expecting approval by the transaction
	var pending *pendingBatch
	for _, pb := range lst {
		if pb.batch.StateTransactionId() == sm.nextStateTransaction.ID() {
			pending = pb
			break
		}
	}
	if pending == nil {
		// for some reason nextStateTransaction references to valid state hash,
		// however the txid of it is different from what is expected by the pending state updates
		sm.log.Errorw("inconsistency: pending batch doesn't contain expected txid",
			"state hash", varStateHash.String(),
			"expected txid", sm.nextStateTransaction.ID().String(),
		)
		return false
	}
	if sm.solidStateValid || sm.solidVariableState == nil {
		if err := pending.nextVariableState.Commit(sm.committee.Address(), pending.batch); err != nil {
			sm.log.Errorw("failed to save state at index #%d", pending.nextVariableState.StateIndex())
			return false
		}
		if sm.solidVariableState != nil {
			sm.log.Infof("STATE TRANSITION %s --> #%d. State hash: %s, state txid: %s",
				sm.solidVariableState.StateIndex(), pending.nextVariableState.StateIndex(),
				varStateHash.String(), sm.nextStateTransaction.ID().String())
		} else {
			sm.log.Infof("ORIGIN STATE SAVED. State hash: %s, state txid: %s",
				varStateHash.String(), sm.nextStateTransaction.ID().String())
		}
	} else {
		// initial load
		sm.log.Infof("STATE #%d CONFIRMED. State hash: %s, state txid: %s",
			sm.solidVariableState.StateIndex(), varStateHash.String(), sm.nextStateTransaction.ID().String())
	}
	sm.solidStateValid = true
	sm.solidVariableState = pending.nextVariableState

	saveTx := sm.nextStateTransaction

	// update state manager variables to the new state
	sm.nextStateTransaction = nil
	sm.pendingBatches = make(map[hashing.HashValue][]*pendingBatch) // clean pending batches
	sm.permutationOfPeers = util.GetPermutation(sm.committee.Size(), varStateHash.Bytes())
	sm.permutationIndex = 0
	sm.syncMessageDeadline = time.Now() // if not synced then immediately

	// if synchronized, notify consensus operator about state transition
	if sm.isSynchronized() {
		// async is a must in order not to deadlock
		go func() {
			sm.committee.ReceiveMessage(&committee.StateTransitionMsg{
				VariableState:    sm.solidVariableState,
				StateTransaction: saveTx,
			})
		}()
	}
	return true
}

const periodBetweenSyncMessages = 1 * time.Second

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
		sm.syncMessageDeadline = time.Now().Add(periodBetweenSyncMessages)
	}
}

const stateTransactionRequestTimeout = 10 * time.Second

func (sm *stateManager) requestStateTransactionIfNeeded() {
	if sm.nextStateTransaction != nil {
		return
	}
	for _, lst := range sm.pendingBatches {
		for _, pb := range lst {
			if pb.stateTransactionRequestDeadline.Before(time.Now()) {
				txid := pb.batch.StateTransactionId()
				sm.log.Debugf("query transaction from the node. txid = %s", txid.String())
				_ = nodeconn.RequestTransactionFromNode(&txid)
				pb.stateTransactionRequestDeadline = time.Now().Add(stateTransactionRequestTimeout)
			}
		}
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
	if sm.solidVariableState == nil {
		return false
	}
	return sm.largestEvidencedStateIndex-sm.solidVariableState.StateIndex() <= 1
}

// adding batch of state updates to the 'pending' map
func (sm *stateManager) addPendingBatch(batch state.Batch) bool {
	if sm.solidStateValid {
		if batch.StateIndex() != sm.solidVariableState.StateIndex()+1 {
			// if current state is validated, only interested in the batches of state updates for the next state
			return false
		}
	} else {
		// initial loading
		if sm.solidVariableState == nil {
			// origin state
			if batch.StateIndex() != 0 {
				sm.log.Errorf("expected batch index 0 got %d", batch.StateIndex())
				return false
			}
		} else {
			// not origin state, the loaded state must be approved by the transaction
			if batch.StateIndex() != sm.solidVariableState.StateIndex() {
				sm.log.Errorf("expected batch index %d got %d",
					sm.solidVariableState.StateIndex(), batch.StateIndex())
				return false
			}
		}
	}
	varStateToApprove := state.NewVariableState(sm.solidVariableState)
	if sm.solidStateValid || sm.solidVariableState == nil {
		// we need to approve the next state.
		// In case of origin, the next state is origin batch applied to empty state
		if err := varStateToApprove.Apply(batch); err != nil {
			sm.log.Errorw("can't apply update to the current state",
				"cur state index", sm.solidVariableState.StateIndex(),
				"err", err,
			)
			return false
		}
	}
	// include the bach to pending batches map
	vh := varStateToApprove.Hash()
	pendingBatches, ok := sm.pendingBatches[vh]
	if !ok {
		pendingBatches = make([]*pendingBatch, 0)
	}
	pb := &pendingBatch{
		batch:             batch,
		nextVariableState: varStateToApprove,
	}
	pendingBatches = append(pendingBatches, pb)
	sm.pendingBatches[vh] = pendingBatches

	// request approving transaction from the node. It may also come without request
	txid := batch.StateTransactionId()
	if err := nodeconn.RequestTransactionFromNode(&txid); err != nil {
		sm.log.Debug(err)
	}
	pb.stateTransactionRequestDeadline = time.Now().Add(stateTransactionRequestTimeout)
	return true
}
