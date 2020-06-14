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
	varStateHash := sm.nextStateTransaction.MustState().VariableStateHash()
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

	if sm.solidStateValid || sm.solidVariableState == nil {
		if sm.solidVariableState == nil {
			// pre-origin
			if sm.nextStateTransaction.ID() != (valuetransaction.ID)(*sm.committee.Color()) {
				sm.log.Errorf("major inconsistency: origin transaction hash %s not equal to the color of the SC %s",
					sm.nextStateTransaction.ID().String(), sm.committee.Color().String())
				sm.committee.Dismiss()
				return false
			}
		}
		if err := pending.nextVariableState.CommitToDb(*sm.committee.Address(), pending.batch); err != nil {
			sm.log.Errorw("failed to save state at index #%d", pending.nextVariableState.StateIndex())
			return false
		}

		//sm.log.Debugf("$$$$$$$$$$$$ COMMITTED STATE\nvarstate: %s\nbatch: %s",
		//	pending.nextVariableState.String(), pending.batch.String())

		if sm.solidVariableState != nil {
			publisher.Publish("state",
				sm.committee.Address().String(),
				fmt.Sprintf("%d", sm.solidVariableState.StateIndex()),
				fmt.Sprintf("%d", pending.nextVariableState.StateIndex()),
				sm.nextStateTransaction.ID().String(),
				varStateHash.String(),
			)

			sm.log.Infof("TRANSITION TO THE NEXT STATE #%d --> #%d. State hash: %s, state txid: %s, batch essence: %s",
				sm.solidVariableState.StateIndex(), pending.nextVariableState.StateIndex(),
				varStateHash.String(), sm.nextStateTransaction.ID().String(), pending.batch.EssenceHash().String())
		} else {
			publisher.Publish("state",
				sm.committee.Address().String(),
				"-1",
				"0",
				sm.nextStateTransaction.ID().String(),
				varStateHash.String(),
			)

			sm.log.Infof("ORIGIN STATE SAVED. State hash: %s, state txid: %s, batch essence: %s",
				varStateHash.String(), sm.nextStateTransaction.ID().String(), pending.batch.EssenceHash().String())
		}

		// publish processed requests
		for _, reqid := range pending.batch.RequestIds() {
			publisher.Publish("request", "out",
				sm.committee.Address().String(),
				reqid.TransactionId().String(),
				fmt.Sprintf("%d", reqid.Index()),
			)
		}
	} else {
		// initial load
		publisher.Publish("state", sm.committee.Address().String(),
			strconv.Itoa(int(sm.solidVariableState.StateIndex())),
			strconv.Itoa(int(sm.solidVariableState.StateIndex())),
			sm.nextStateTransaction.ID().String(),
			varStateHash.String(),
		)

		sm.log.Infof("INITIAL STATE #%d LOADED. State hash: %s, state txid: %s",
			sm.solidVariableState.StateIndex(), varStateHash.String(), sm.nextStateTransaction.ID().String())
	}
	sm.solidStateValid = true
	sm.solidVariableState = pending.nextVariableState

	saveTx := sm.nextStateTransaction

	// update state manager variables to the new state
	sm.nextStateTransaction = nil
	sm.pendingBatches = make(map[hashing.HashValue]*pendingBatch) // clear pending batches
	sm.permutation.Shuffle(varStateHash.Bytes())
	sm.syncMessageDeadline = time.Now() // if not synced then immediately

	go func() {
		sm.committee.ReceiveMessage(&committee.StateTransitionMsg{
			VariableState:    sm.solidVariableState,
			StateTransaction: saveTx,
			Synchronized:     sm.isSynchronized(),
		})
	}()
	return true
}

const periodBetweenSyncMessages = 1 * time.Second

func (sm *stateManager) requestStateUpdateFromPeerIfNeeded() {
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
	stateIndex := uint32(0)
	if sm.solidVariableState != nil {
		stateIndex = sm.solidVariableState.StateIndex() + 1
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
		sm.syncMessageDeadline = time.Now().Add(periodBetweenSyncMessages)
	}
}

// index of evidenced state index is passed to record the largest one.
// This is needed to check synchronization status.
func (sm *stateManager) EvidenceStateIndex(stateIndex uint32) {
	// synced state is when current state index is behind
	// the largestEvidencedStateIndex no more than by 1 point
	wasSynchronized := sm.isSynchronized()

	currStateIndex := int32(-1)
	if sm.solidVariableState != nil {
		currStateIndex = int32(sm.solidVariableState.StateIndex())
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
		sm.log.Debugf("SYNCED: current state index: %d", sm.solidVariableState.StateIndex())
	}
}

func (sm *stateManager) isSynchronized() bool {
	if sm.solidVariableState == nil {
		return false
	}
	return sm.largestEvidencedStateIndex == sm.solidVariableState.StateIndex()
}

var niltxid valuetransaction.ID

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
	varStateToApprove := state.NewVariableState(sm.solidVariableState) // clone

	if sm.solidStateValid || sm.solidVariableState == nil {
		// we need to approve the solidVariableState.
		// In case of origin, the next state is origin batch applied to the empty state
		if err := varStateToApprove.ApplyBatch(batch); err != nil {
			sm.log.Errorw("can't apply update to the current state",
				"cur state index", sm.solidVariableState.StateIndex(),
				"err", err,
			)
			return false
		}
	}
	// include the bach to pending batches map
	vh := varStateToApprove.Hash()
	pb, ok := sm.pendingBatches[vh]
	if !ok || pb.batch.StateTransactionId() == niltxid {
		pb = &pendingBatch{
			batch:             batch,
			nextVariableState: varStateToApprove,
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

const stateTransactionRequestTimeout = 10 * time.Second

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
	pb.stateTransactionRequestDeadline = time.Now().Add(stateTransactionRequestTimeout)
}
