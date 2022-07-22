// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package statemgr

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/chain/messages"
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
		sm.log.Debugf("handleGetBlockMsg ignored: current state output index #%d is older than requested block index #%d",
			sm.stateOutput.GetStateIndex(), msg.BlockIndex)
		return
	}
	blockBytes, err := state.LoadBlockBytes(sm.store, msg.BlockIndex)
	if err != nil {
		sm.log.Errorf("handleGetBlockMsg: LoadBlockBytes error: %v", err)
		return
	}
	if blockBytes == nil {
		sm.log.Debugf("handleGetBlockMsg ignored: block index #%d not found", msg.BlockIndex)
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

func (sm *stateManager) EnqueueAliasOutput(output *iscp.AliasOutputWithID) {
	sm.eventAliasOutputPipe.In() <- output
}

func (sm *stateManager) handleAliasOutput(output *iscp.AliasOutputWithID) {
	sm.log.Debugf("EventAliasOutput received: output id %s for state index %v", iscp.OID(output.ID()), output.GetStateIndex())
	// sm.stateManagerMetrics.LastSeenStateIndex(msg.ChainOutput.GetStateIndex()) //TODO!!!
	stateL1Commitment, err := state.L1CommitmentFromAliasOutput(output.GetAliasOutput())
	if err != nil {
		sm.log.Errorf("EventAliasOutput ignored: failed to parse state commitment: %v", err)
		return
	}
	sm.log.Debugf("EventAliasOutput received: state commitment is %s", stateL1Commitment.StateCommitment)
	if sm.aliasOutputReceived(output) {
		sm.takeAction()
	}
}

func (sm *stateManager) EnqueueStateCandidateMsg(virtualState state.VirtualStateAccess, outputID *iotago.UTXOInput) {
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
