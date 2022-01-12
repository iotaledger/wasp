// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package statemgr

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/chain/messages"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util"
)

// EventGetBlockMsg is a request for a block while syncing
func (sm *stateManager) EnqueueGetBlockMsg(msg *messages.GetBlockMsgIn) {
	sm.eventGetBlockMsgPipe.In() <- msg
}

func (sm *stateManager) handleGetBlockMsg(msg *messages.GetBlockMsgIn) {
	sm.log.Debugw("handleGetBlockMsg: ",
		"sender", msg.SenderPubKey.String(),
		"block index", msg.BlockIndex,
	)
	if sm.stateOutput == nil { // Not a necessary check, only for optimization.
		sm.log.Debugf("handleGetBlockMsg: message ignored: stateOutput is nil")
		return
	}
	if msg.BlockIndex > sm.stateOutput.GetStateIndex() { // Not a necessary check, only for optimization.
		sm.log.Debugf("handleGetBlockMsg: message ignored 1: block #%d not found. Current state index: #%d",
			msg.BlockIndex, sm.stateOutput.GetStateIndex())
		return
	}
	blockBytes, err := state.LoadBlockBytes(sm.store, msg.BlockIndex)
	if err != nil {
		sm.log.Errorf("handleGetBlockMsg: LoadBlockBytes error: %v", err)
		return
	}
	if blockBytes == nil {
		sm.log.Debugf("handleGetBlockMsg message ignored 2: block #%d not found. Current state index: #%d",
			msg.BlockIndex, sm.stateOutput.GetStateIndex())
		return
	}

	sm.log.Debugf("handleGetBlockMsg: responding to peer %s by block %v", msg.SenderPubKey.String(), msg.BlockIndex)

	blockMsg := &messages.BlockMsg{BlockBytes: blockBytes}
	sm.domain.SendMsgByPubKey(msg.SenderPubKey, peering.PeerMessageReceiverStateManager, peerMsgTypeBlock, util.MustBytes(blockMsg))
}

// EventBlockMsg
func (sm *stateManager) EnqueueBlockMsg(msg *messages.BlockMsgIn) {
	sm.eventBlockMsgPipe.In() <- msg
}

func (sm *stateManager) handleBlockMsg(msg *messages.BlockMsgIn) {
	sm.syncingBlocks.blockReceived()
	sm.log.Debugw("handleBlockMsg: ",
		"sender", msg.SenderPubKey.String(),
	)
	if sm.stateOutput == nil {
		sm.log.Debugf("handleBlockMsg: message ignored: stateOutput is nil")
		return
	}
	block, err := state.BlockFromBytes(msg.BlockBytes)
	if err != nil {
		sm.log.Warnf("handleBlockMsg: message ignored: wrong block received from peer %s. Err: %v", msg.SenderPubKey.String(), err)
		return
	}
	sm.log.Debugw("handleBlockMsg: adding block from peer ",
		"sender", msg.SenderPubKey.String(),
		"block index", block.BlockIndex(),
		"approving output", iscp.OID(block.ApprovingOutputID()),
	)
	if sm.addBlockFromPeer(block) {
		sm.takeAction()
	}
}

func (sm *stateManager) EnqueueOutputMsg(msg iotago.Output) {
	sm.eventOutputMsgPipe.In() <- msg
}

func (sm *stateManager) handleOutputMsg(msg iotago.Output) {
	sm.log.Debugf("EventOutputMsg received: %s", iscp.OID(msg.ID()))
	chainOutput, ok := msg.(*iotago.AliasOutput)
	if !ok {
		sm.log.Debugf("EventOutputMsg ignored: output is of type %t, expecting *iotago.AliasOutput", msg)
		return
	}
	if sm.outputPulled(chainOutput) {
		sm.takeAction()
	}
}

// EventStateTransactionMsg triggered whenever new state transaction arrives
// the state transaction may be confirmed or not
func (sm *stateManager) EnqueueStateMsg(msg *messages.StateMsg) {
	sm.eventStateOutputMsgPipe.In() <- msg
}

func (sm *stateManager) handleStateMsg(msg *messages.StateMsg) {
	sm.log.Debugw("EventStateMsg received: ",
		"state index", msg.ChainOutput.GetStateIndex(),
		"chainOutput", iscp.OID(msg.ChainOutput.ID()),
	)
	stateHash, err := hashing.HashValueFromBytes(msg.ChainOutput.GetStateData())
	if err != nil {
		sm.log.Errorf("EventStateMsg ignored: failed to parse state hash: %v", err)
		return
	}
	sm.log.Debugf("EventStateMsg state hash is %v", stateHash.String())
	if sm.stateOutputReceived(msg.ChainOutput, msg.Timestamp) {
		sm.takeAction()
	}
}

func (sm *stateManager) EnqueueStateCandidateMsg(virtualState state.VirtualStateAccess, outputID iotago.OutputID) {
	sm.eventStateCandidateMsgPipe.In() <- &messages.StateCandidateMsg{
		State:             virtualState,
		ApprovingOutputID: outputID,
	}
}

func (sm *stateManager) handleStateCandidateMsg(msg *messages.StateCandidateMsg) {
	sm.log.Debugf("EventStateCandidateMsg received: state index: %d, timestamp: %v",
		msg.State.BlockIndex(), msg.State.Timestamp(),
	)
	if sm.stateOutput == nil {
		sm.log.Debugf("EventStateCandidateMsg ignored: stateOutput is nil")
		return
	}
	if sm.addStateCandidateFromConsensus(msg.State, msg.ApprovingOutputID) {
		sm.takeAction()
	}
}

func (sm *stateManager) EnqueueTimerMsg(msg messages.TimerTick) {
	if msg%2 == 0 {
		sm.eventTimerMsgPipe.In() <- msg
	}
}

func (sm *stateManager) handleTimerMsg() {
	sm.log.Debugf("EventTimerMsg received")
	sm.takeAction()
}
