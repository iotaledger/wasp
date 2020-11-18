package statemgr

import (
	"fmt"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/plugins/nodeconn"
	"github.com/iotaledger/wasp/plugins/publisher"
	"strconv"
	"time"
)

func (sm *stateManager) takeAction() {
	sm.sendPingsIfNeeded()
	sm.notifyConsensusOnStateTransitionIfNeeded()
	if sm.checkStateApproval() {
		return
	}
	sm.requestStateTransactionIfNeeded()
	sm.requestStateUpdateFromPeerIfNeeded()
}

func (sm *stateManager) notifyConsensusOnStateTransitionIfNeeded() {
	if !sm.solidStateValid {
		return
	}
	if sm.consensusNotifiedOnStateTransition {
		return
	}
	if !sm.numPongsHasQuorum() {
		return
	}

	sm.consensusNotifiedOnStateTransition = true
	go sm.chain.ReceiveMessage(&chain.StateTransitionMsg{
		VariableState:     sm.solidState,
		AnchorTransaction: sm.approvingTransaction,
		Synchronized:      sm.isSynchronized(),
	})
}

// sendPingsIfNeeded sends pings to the committee peers to gather evidence about the largest
// state index. It doesn't send it if evidence is already here.
// To force sending pings all 'false' must be set in sm.pingPong
func (sm *stateManager) sendPingsIfNeeded() {
	if sm.numPongsHasQuorum() {
		// no need for pinging, all state information is gathered already
		return
	}
	if !sm.chain.HasQuorum() {
		// doesn't make sense the pinging, alive nodes does not make quorum
		return
	}
	if !sm.solidStateValid {
		// own solid state has not been validated yet
		return
	}
	if sm.deadlineForPongQuorum.After(time.Now()) {
		// not time yet
		return
	}
	sm.sendPingsToCommitteePeers()
}

// checks the state of the state manager. If one of pending blocks is confirmed
// by the nextStateTransaction changes the state to the next
func (sm *stateManager) checkStateApproval() bool {
	if sm.nextStateTransaction == nil {
		return false
	}
	// among pending state update batches we locate the one which
	// is approved by the transaction
	varStateHash := sm.nextStateTransaction.MustState().StateHash()
	pending, ok := sm.pendingBlocks[varStateHash]
	if !ok {
		// corresponding block wasn't found among pending state updates
		// transaction doesn't approve anything
		return false
	}

	// found a pending block which is approved by the nextStateTransaction

	if pending.block.StateTransactionID() == niltxid {
		// not committed yet block. Link it to the transaction
		pending.block.WithStateTransaction(sm.nextStateTransaction.ID())
	} else {
		txid1 := pending.block.StateTransactionID()
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
			if sm.nextStateTransaction.ID() != (valuetransaction.ID)(*sm.chain.Color()) {
				sm.log.Errorf("major inconsistency: origin transaction hash %s not equal to the color of the SC %s",
					sm.nextStateTransaction.ID().String(), sm.chain.Color().String())
				sm.chain.Dismiss()
				return false
			}
		}
		if err := pending.nextState.CommitToDb(pending.block); err != nil {
			sm.log.Errorw("failed to save state at index #%d", pending.nextState.BlockIndex())
			return false
		}

		if sm.solidState != nil {
			sm.log.Infof("STATE TRANSITION TO #%d. Anchor transaction: %s, block size: %d",
				pending.nextState.BlockIndex(), sm.nextStateTransaction.ID().String(), pending.block.Size())
			sm.log.Debugf("STATE TRANSITION. State hash: %s, block essence: %s",
				varStateHash.String(), pending.block.EssenceHash().String())
		} else {
			sm.log.Infof("ORIGIN STATE SAVED. Origin transaction: %s",
				sm.nextStateTransaction.ID().String())
			sm.log.Debugf("ORIGIN STATE SAVED. State hash: %s, state txid: %s, block essence: %s",
				varStateHash.String(), sm.nextStateTransaction.ID().String(), pending.block.EssenceHash().String())
		}

	} else {
		// !sm.solidStateValid && sm.solidState != nil --> initial load

		sm.log.Infof("INITIAL STATE #%d LOADED FROM DB. State hash: %s, state txid: %s",
			sm.solidState.BlockIndex(), varStateHash.String(), sm.nextStateTransaction.ID().String())
	}
	sm.solidStateValid = true
	sm.solidState = pending.nextState

	sm.approvingTransaction = sm.nextStateTransaction

	// update state manager variables to the new state
	sm.nextStateTransaction = nil
	sm.pendingBlocks = make(map[hashing.HashValue]*pendingBlock) // clear pending batches
	sm.permutation.Shuffle(varStateHash.Bytes())
	sm.syncMessageDeadline = time.Now() // if not synced then immediately
	sm.consensusNotifiedOnStateTransition = false

	// publish state transition
	publisher.Publish("state",
		sm.chain.ID().String(),
		strconv.Itoa(int(sm.solidState.BlockIndex())),
		strconv.Itoa(int(pending.block.Size())),
		sm.approvingTransaction.ID().String(),
		varStateHash.String(),
		fmt.Sprintf("%d", pending.block.Timestamp()),
	)
	// publish processed requests
	for i, reqid := range pending.block.RequestIDs() {

		sm.chain.EventRequestProcessed().Trigger(*reqid)

		publisher.Publish("request_out",
			sm.chain.ID().String(),
			reqid.TransactionID().String(),
			fmt.Sprintf("%d", reqid.Index()),
			strconv.Itoa(int(sm.solidState.BlockIndex())),
			strconv.Itoa(i),
			strconv.Itoa(int(pending.block.Size())),
		)
	}
	return true
}

func (sm *stateManager) requestStateUpdateFromPeerIfNeeded() {
	if !sm.solidStateValid || sm.isSynchronized() {
		// no need for more info when state is synced or solid state still needs validation by the anchor tx
		return
	}
	// state is valid but not synced
	if !sm.syncMessageDeadline.Before(time.Now()) {
		// not time yet for the next message
		return
	}
	// it is time to ask for the next state update to next peer in the permutation
	data := util.MustBytes(&chain.GetBlockMsg{
		PeerMsgHeader: chain.PeerMsgHeader{
			BlockIndex: sm.solidState.BlockIndex() + 1,
		},
	})
	// send messages until first without error
	for i := uint16(0); i < sm.chain.Size(); i++ {
		if err := sm.chain.SendMsg(sm.permutation.Next(), chain.MsgGetBatch, data); err == nil {
			break
		}
		sm.syncMessageDeadline = time.Now().Add(chain.PeriodBetweenSyncMessages)
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
		currStateIndex = int32(sm.solidState.BlockIndex())
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
		sm.log.Debugf("SYNCED: current state index: %d", sm.solidState.BlockIndex())
	}
}

func (sm *stateManager) isSynchronized() bool {
	if sm.solidState == nil {
		return false // sm.largestEvidencedStateIndex == 0
	}
	return sm.largestEvidencedStateIndex == sm.solidState.BlockIndex()
}

var niltxid valuetransaction.ID

// adding block of state updates to the 'pending' map
func (sm *stateManager) addPendingBlock(block state.Block) bool {
	sm.log.Debugw("addPendingBlock",
		"block index", block.StateIndex(),
		"timestamp", block.Timestamp(),
		"size", block.Size(),
		"state tx", block.StateTransactionID().String(),
	)

	if sm.solidStateValid {
		if block.StateIndex() != sm.solidState.BlockIndex()+1 {
			// if current state is validated, only interested in the batches of state updates for the next state
			return false
		}
	} else {
		// initial loading
		if sm.solidState == nil {
			// origin state
			if block.StateIndex() != 0 {
				sm.log.Errorf("addPendingBlock: expected block index 0 got %d", block.StateIndex())
				return false
			}
		} else {
			// not origin state, the loaded state must be approved by the transaction
			if block.StateIndex() != sm.solidState.BlockIndex() {
				sm.log.Errorf("addPendingBlock: expected block index %d got %d",
					sm.solidState.BlockIndex(), block.StateIndex())
				return false
			}
		}
	}

	stateToApprove := sm.createStateToApprove()

	if sm.solidStateValid || sm.solidState == nil {
		// we need to approve the solidState.
		// In case of origin, the next state is origin block applied to the empty state
		if err := stateToApprove.ApplyBatch(block); err != nil {
			sm.log.Errorw("can't apply update to the current state",
				"cur state index", sm.solidState.BlockIndex(),
				"err", err,
			)
			return false
		}
	}

	// include the bach to pending batches map
	vh := stateToApprove.Hash()
	pb, ok := sm.pendingBlocks[*vh]
	if !ok || pb.block.StateTransactionID() == niltxid {
		pb = &pendingBlock{
			block:     block,
			nextState: stateToApprove,
		}
		sm.pendingBlocks[*vh] = pb
	}

	sm.log.Debugw("added new pending block",
		"state index", pb.block.StateIndex(),
		"state hash", vh.String(),
		"approving tx", pb.block.StateTransactionID().String(),
	)
	// request approving transaction from the node. It may also come without request
	if block.StateTransactionID() != niltxid {
		sm.requestStateTransaction(pb)
	}
	return true
}

func (sm *stateManager) createStateToApprove() state.VirtualState {
	if sm.solidState == nil {
		return state.NewEmptyVirtualState(sm.chain.ID())
	}
	return sm.solidState.Clone()
}

// for committed batches request approving transaction if deadline has passed
func (sm *stateManager) requestStateTransactionIfNeeded() {
	if sm.nextStateTransaction != nil {
		return
	}
	for _, pb := range sm.pendingBlocks {
		if pb.block.StateTransactionID() != niltxid && pb.stateTransactionRequestDeadline.Before(time.Now()) {
			sm.requestStateTransaction(pb)
		}
	}
}

func (sm *stateManager) requestStateTransaction(pb *pendingBlock) {
	txid := pb.block.StateTransactionID()
	sm.log.Debugf("query transaction from the node. txid = %s", txid.String())
	_ = nodeconn.RequestConfirmedTransactionFromNode(&txid)
	pb.stateTransactionRequestDeadline = time.Now().Add(chain.StateTransactionRequestTimeout)
}

func (sm *stateManager) numPongs() uint16 {
	ret := uint16(0)
	for _, f := range sm.pingPong {
		if f {
			ret++
		}
	}
	return ret
}

func (sm *stateManager) numPongsHasQuorum() bool {
	return sm.numPongs() >= sm.chain.Quorum()-1
}

func (sm *stateManager) pingPongReceived(senderIndex uint16) {
	sm.pingPong[senderIndex] = true
}

func (sm *stateManager) respondPongToPeer(targetPeerIndex uint16) {
	sm.chain.SendMsg(targetPeerIndex, chain.MsgStateIndexPingPong, util.MustBytes(&chain.StateIndexPingPongMsg{
		PeerMsgHeader: chain.PeerMsgHeader{
			BlockIndex: sm.solidState.BlockIndex(),
		},
		RSVP: false,
	}))
}

func (sm *stateManager) sendPingsToCommitteePeers() {
	sm.log.Debugf("pinging peers")

	data := util.MustBytes(&chain.StateIndexPingPongMsg{
		PeerMsgHeader: chain.PeerMsgHeader{
			BlockIndex: sm.solidState.BlockIndex(),
		},
		RSVP: true,
	})
	numSent := 0
	for i, pinged := range sm.pingPong {
		if pinged {
			continue
		}
		if err := sm.chain.SendMsg(uint16(i), chain.MsgStateIndexPingPong, data); err == nil {
			numSent++
		}
	}
	sm.log.Debugf("sent pings to %d committee peers", numSent)
	sm.deadlineForPongQuorum = time.Now().Add(chain.RepeatPingAfter)
}
