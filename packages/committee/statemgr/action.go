package statemgr

import (
	"fmt"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/packages/committee"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/plugins/nodeconn"
	"github.com/iotaledger/wasp/plugins/publisher"
	"strconv"
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
	varStateHash := sm.nextStateTransaction.MustState().StateHash()
	pending, ok := sm.pendingBatches[varStateHash]
	if !ok {
		// corresponding batch wasn't found among pending state updates
		// transaction doesn't approve anything
		return false
	}

	// found a pending batch which is approved by the nextStateTransaction

	if pending.batch.StateTransactionId() == niltxid {
		// not committed yet batch. Link it to the transaction
		pending.batch.WithStateTransaction(sm.nextStateTransaction.ID())
	} else {
		txid1 := pending.batch.StateTransactionId()
		txid2 := sm.nextStateTransaction.ID()
		if txid1 != txid2 {
			sm.log.Errorf("major inconsistency: var state hash %s is approved by two different tx: txid1 = %s, txid2 = %s",
				varStateHash.String(), txid1.String(), txid2.String())
			return false
		}
	}

	if sm.solidStateValid || sm.solidState == nil {
		if sm.solidState == nil {
			// pre-origin
			if sm.nextStateTransaction.ID() != (valuetransaction.ID)(*sm.committee.Color()) {
				sm.log.Errorf("major inconsistency: origin transaction hash %s not equal to the color of the SC %s",
					sm.nextStateTransaction.ID().String(), sm.committee.Color().String())
				sm.committee.Dismiss()
				return false
			}
		}
		if err := pending.nextState.CommitToDb(pending.batch); err != nil {
			sm.log.Errorw("failed to save state at index #%d", pending.nextState.StateIndex())
			return false
		}

		if sm.solidState != nil {
			sm.log.Infof("STATE TRANSITION TO #%d. Anchor transaction: %s, batch size: %d",
				pending.nextState.StateIndex(), sm.nextStateTransaction.ID().String(), pending.batch.Size())
			sm.log.Debugf("STATE TRANSITION. AccessState hash: %s, batch essence: %s",
				varStateHash.String(), pending.batch.EssenceHash().String())
		} else {
			sm.log.Infof("ORIGIN STATE SAVED. Origin transaction: %s",
				sm.nextStateTransaction.ID().String())
			sm.log.Debugf("ORIGIN STATE SAVED. AccessState hash: %s, state txid: %s, batch essence: %s",
				varStateHash.String(), sm.nextStateTransaction.ID().String(), pending.batch.EssenceHash().String())
		}

	} else {
		// initial load

		sm.log.Infof("INITIAL STATE #%d LOADED. AccessState hash: %s, state txid: %s",
			sm.solidState.StateIndex(), varStateHash.String(), sm.nextStateTransaction.ID().String())
	}
	sm.solidStateValid = true
	sm.solidState = pending.nextState

	saveTx := sm.nextStateTransaction

	// update state manager variables to the new state
	sm.nextStateTransaction = nil
	sm.pendingBatches = make(map[hashing.HashValue]*pendingBatch) // clear pending batches
	sm.permutation.Shuffle(varStateHash.Bytes())
	sm.syncMessageDeadline = time.Now() // if not synced then immediately

	// publish state transition
	publisher.Publish("state",
		sm.committee.Address().String(),
		strconv.Itoa(int(sm.solidState.StateIndex())),
		strconv.Itoa(int(pending.batch.Size())),
		saveTx.ID().String(),
		varStateHash.String(),
		fmt.Sprintf("%d", pending.batch.Timestamp()),
	)
	// publish processed requests
	for i, reqid := range pending.batch.RequestIds() {
		publisher.Publish("request_out",
			sm.committee.Address().String(),
			reqid.TransactionId().String(),
			fmt.Sprintf("%d", reqid.Index()),
			strconv.Itoa(int(sm.solidState.StateIndex())),
			strconv.Itoa(i),
			strconv.Itoa(int(pending.batch.Size())),
		)
	}

	go func() {
		sm.committee.ReceiveMessage(&committee.StateTransitionMsg{
			VariableState:    sm.solidState,
			StateTransaction: saveTx,
			Synchronized:     sm.isSynchronized(),
		})
	}()
	return true
}

func (sm *stateManager) requestStateUpdateFromPeerIfNeeded() {
	if sm.isSynchronized() || sm.solidState == nil {
		// state is synced, no need for more info
		// or it is in the pre-origin state, the 0 batch is deterministically known
		return
	}
	// not synced
	if !sm.syncMessageDeadline.Before(time.Now()) {
		// not time yet for the next message
		return
	}
	// it is time to ask for the next state update to next peer in the permutation
	stateIndex := uint32(0)
	if sm.solidState != nil {
		stateIndex = sm.solidState.StateIndex() + 1
	}
	data := util.MustBytes(&committee.GetBatchMsg{
		PeerMsgHeader: committee.PeerMsgHeader{
			StateIndex: stateIndex,
		},
	})
	// send messages until first without error
	for i := uint16(0); i < sm.committee.Size(); i++ {
		if err := sm.committee.SendMsg(sm.permutation.Next(), committee.MsgGetBatch, data); err == nil {
			break
		}
		sm.syncMessageDeadline = time.Now().Add(committee.PeriodBetweenSyncMessages)
	}
}

// index of evidenced state index is passed to record the largest one.
// This is needed to check synchronization status.
func (sm *stateManager) EvidenceStateIndex(stateIndex uint32) {
	// synced state is when current state index is behind
	// the largestEvidencedStateIndex no more than by 1 point
	wasSynchronized := sm.isSynchronized()

	currStateIndex := int32(-1)
	if sm.solidState != nil {
		currStateIndex = int32(sm.solidState.StateIndex())
	}

	if stateIndex > sm.largestEvidencedStateIndex {
		sm.largestEvidencedStateIndex = stateIndex
	}
	switch {
	case !sm.isSynchronized() && wasSynchronized:
		sm.syncMessageDeadline = time.Now()
		sm.log.Debugf("NOT SYNCED: current state index: %d, largest evidenced index: %d",
			currStateIndex, sm.largestEvidencedStateIndex)
	case sm.isSynchronized() && !wasSynchronized:
		sm.log.Debugf("SYNCED: current state index: %d", sm.solidState.StateIndex())
	}
}

func (sm *stateManager) isSynchronized() bool {
	if sm.solidState == nil {
		return sm.largestEvidencedStateIndex == 0
	}
	return sm.largestEvidencedStateIndex == sm.solidState.StateIndex()
}

var niltxid valuetransaction.ID

// adding batch of state updates to the 'pending' map
func (sm *stateManager) addPendingBatch(batch state.Batch) bool {
	sm.log.Debugw("addPendingBatch",
		"state index", batch.StateIndex(),
		"timestamp", batch.Timestamp(),
		"size", batch.Size(),
		"state tx", batch.StateTransactionId().String(),
	)

	if sm.solidStateValid {
		if batch.StateIndex() != sm.solidState.StateIndex()+1 {
			// if current state is validated, only interested in the batches of state updates for the next state
			return false
		}
	} else {
		// initial loading
		if sm.solidState == nil {
			// origin state
			if batch.StateIndex() != 0 {
				sm.log.Errorf("expected batch index 0 got %d", batch.StateIndex())
				return false
			}
		} else {
			// not origin state, the loaded state must be approved by the transaction
			if batch.StateIndex() != sm.solidState.StateIndex() {
				sm.log.Errorf("expected batch index %d got %d",
					sm.solidState.StateIndex(), batch.StateIndex())
				return false
			}
		}
	}

	stateToApprove := sm.createStateToApprove()

	if sm.solidStateValid || sm.solidState == nil {
		// we need to approve the solidState.
		// In case of origin, the next state is origin batch applied to the empty state
		if err := stateToApprove.ApplyBatch(batch); err != nil {
			sm.log.Errorw("can't apply update to the current state",
				"cur state index", sm.solidState.StateIndex(),
				"err", err,
			)
			return false
		}
	}

	// include the bach to pending batches map
	vh := stateToApprove.Hash()
	pb, ok := sm.pendingBatches[vh]
	if !ok || pb.batch.StateTransactionId() == niltxid {
		pb = &pendingBatch{
			batch:     batch,
			nextState: stateToApprove,
		}
		sm.pendingBatches[vh] = pb
	}

	sm.log.Debugw("added new pending batch",
		"state index", pb.batch.StateIndex(),
		"state hash", vh.String(),
		"approving tx", pb.batch.StateTransactionId().String(),
	)
	// request approving transaction from the node. It may also come without request
	if batch.StateTransactionId() != niltxid {
		sm.requestStateTransaction(pb)
	}
	return true
}

func (sm *stateManager) createStateToApprove() state.VirtualState {
	if sm.solidState == nil {
		return state.NewEmptyVirtualState(sm.committee.Address())
	}
	return sm.solidState.Clone()
}

// for committed batches request approving transaction if deadline has passed
func (sm *stateManager) requestStateTransactionIfNeeded() {
	if sm.nextStateTransaction != nil {
		return
	}
	for _, pb := range sm.pendingBatches {
		if pb.batch.StateTransactionId() != niltxid && pb.stateTransactionRequestDeadline.Before(time.Now()) {
			sm.requestStateTransaction(pb)
		}
	}
}

func (sm *stateManager) requestStateTransaction(pb *pendingBatch) {
	txid := pb.batch.StateTransactionId()
	sm.log.Debugf("query transaction from the node. txid = %s", txid.String())
	_ = nodeconn.RequestTransactionFromNode(&txid)
	pb.stateTransactionRequestDeadline = time.Now().Add(committee.StateTransactionRequestTimeout)
}
