// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package statemgr

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/coretypes"
	"strconv"
	"time"

	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/publisher"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util"
)

func (sm *stateManager) takeAction() {
	sm.sendPingsIfNeeded()
	sm.pullStateIfNeeded()
	sm.doSyncIfNeeded()
	sm.notifyConsensusOnStateTransitionIfNeeded()
}

func (sm *stateManager) pullStateIfNeeded() {
	nowis := time.Now()
	if sm.pullStateDeadline.Before(nowis) {
		return
	}
	if sm.stateOutput == nil || len(sm.pendingBlocks) > 0 {
		sm.nodeConn.RequestBacklog(sm.chain.ID().AsAddress())
	}
	sm.pullStateDeadline = nowis.Add(pullStateTimeout)
}

func (sm *stateManager) sendPingsIfNeeded() {
	if sm.numPongsHasQuorum() {
		// no need for pinging, all state information is gathered already
		return
	}
	if !sm.peers.NumIsAlive(sm.peers.NumPeers()/3 + 1) {
		return
	}
	if !sm.isSolidStateValidated() {
		// own solid state has not been validated yet
		return
	}
	if sm.deadlineForPongQuorum.After(time.Now()) {
		// not time yet
		return
	}
	sm.sendPingsToPeers()
}

func (sm *stateManager) isSolidStateValidated() bool {
	if sm.stateOutput == nil {
		return false
	}
	if sm.solidState == nil && sm.stateOutput.GetStateIndex() == 0 {
		return true
	}
	if sm.solidState != nil && sm.stateOutput.GetStateIndex() == sm.solidState.BlockIndex() {
		return true
	}
	return false
}

func (sm *stateManager) notifyConsensusOnStateTransitionIfNeeded() {
	if sm.stateOutput == nil {
		return
	}
	if !sm.isSolidStateValidated() {
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
		VariableState: sm.solidState.Clone(),
		ChainOutput:   sm.stateOutput,
		Timestamp:     sm.stateOutputTimestamp,
	})
}

func (sm *stateManager) checkStateApproval() {
	if !sm.isSolidStateValidated() {
		return
	}
	// among pending state update batches we locate the one which
	// is approved by the state output
	varStateHash, err := hashing.HashValueFromBytes(sm.stateOutput.GetStateData())
	if err != nil {
		sm.log.Panic(err)
	}
	pending, ok := sm.pendingBlocks[varStateHash]
	if !ok {
		// corresponding block wasn't found among pending state updates
		// transaction doesn't approve anything
		return
	}

	// found a pending block which is approved by the stateOutput
	// set the transaction id from output
	pending.block.WithApprovingOutputID(sm.stateOutput.ID())

	if err := pending.nextState.CommitToDb(pending.block); err != nil {
		sm.log.Errorw("failed to save state at index #%d", pending.nextState.BlockIndex())
		return
	}
	if sm.solidState != nil {
		sm.log.Infof("STATE TRANSITION TO #%d. Chain output: %s, block size: %d",
			pending.nextState.BlockIndex(), coretypes.OID(sm.stateOutput.ID()), pending.block.Size())
		sm.log.Debugf("STATE TRANSITION. State hash: %s, block essence: %s",
			varStateHash.String(), pending.block.EssenceHash().String())
	} else {
		sm.log.Infof("ORIGIN STATE SAVED. State output id: %s", coretypes.OID(sm.stateOutput.ID()))
		sm.log.Debugf("ORIGIN STATE SAVED. state hash: %s, block essence: %s",
			varStateHash.String(), pending.block.EssenceHash().String())
	}
	sm.solidState = pending.nextState
	sm.pendingBlocks = make(map[hashing.HashValue]*pendingBlock) // clear pending batches
	sm.consensusNotifiedOnStateTransition = false

	// publish state transition
	publisher.Publish("state",
		sm.chain.ID().String(),
		strconv.Itoa(int(sm.solidState.BlockIndex())),
		strconv.Itoa(int(pending.block.Size())),
		sm.stateOutput.ID().String(),
		varStateHash.String(),
		fmt.Sprintf("%d", pending.block.Timestamp()),
	)
	// publish processed requests
	for i, reqid := range pending.block.RequestIDs() {

		sm.chain.EventRequestProcessed().Trigger(reqid)

		publisher.Publish("request_out",
			sm.chain.ID().String(),
			reqid.String(),
			strconv.Itoa(int(sm.solidState.BlockIndex())),
			strconv.Itoa(i),
			strconv.Itoa(int(pending.block.Size())),
		)
	}
}

// adding block of state updates to the 'pending' map
func (sm *stateManager) addPendingBlock(block state.Block) {
	sm.log.Debugw("addPendingBlock",
		"block index", block.StateIndex(),
		"timestamp", block.Timestamp(),
		"size", block.Size(),
		"approving output", coretypes.OID(block.ApprovingOutputID()),
	)

	var stateToApprove state.VirtualState
	if sm.solidState == nil {
		stateToApprove = state.NewEmptyVirtualState(sm.chain.ID())
	} else {
		stateToApprove = sm.solidState.Clone()
	}
	if err := stateToApprove.ApplyBlock(block); err != nil {
		sm.log.Error("can't apply update to the current state: %v", err)
		return
	}
	// include the bach to pending batches map
	vh := stateToApprove.Hash()
	sm.pendingBlocks[vh] = &pendingBlock{
		block:     block,
		nextState: stateToApprove,
	}

	sm.log.Debugf("added new pending block. State index: %d, state hash: %s", block.StateIndex(), vh.String())
	sm.pullStateDeadline = time.Now().Add(pullStateTimeout)
}

func (sm *stateManager) createStateToApprove() state.VirtualState {
	if sm.solidState == nil {
		return state.NewEmptyVirtualState(sm.chain.ID())
	}
	return sm.solidState.Clone()
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
	return sm.numPongs() >= sm.peers.NumPeers()/3
}

func (sm *stateManager) pingPongReceived(senderIndex uint16) {
	sm.pingPong[senderIndex] = true
}

func (sm *stateManager) respondPongToPeer(targetPeerIndex uint16) {
	_ = sm.peers.SendMsg(targetPeerIndex, chain.MsgStateIndexPingPong, util.MustBytes(&chain.BlockIndexPingPongMsg{
		BlockIndex: sm.stateOutput.GetStateIndex(),
		RSVP:       false,
	}))
}

func (sm *stateManager) sendPingsToPeers() {
	sm.log.Debugf("pinging peers")

	data := util.MustBytes(&chain.BlockIndexPingPongMsg{
		BlockIndex: sm.stateOutput.GetStateIndex(),
		RSVP:       true,
	})
	numSent := 0
	for i, pinged := range sm.pingPong {
		if pinged {
			continue
		}
		if err := sm.peers.SendMsg(uint16(i), chain.MsgStateIndexPingPong, data); err == nil {
			numSent++
		}
	}
	sm.log.Debugf("sent pings to %d committee peers", numSent)
	sm.deadlineForPongQuorum = time.Now().Add(chain.RepeatPingAfter)
}
