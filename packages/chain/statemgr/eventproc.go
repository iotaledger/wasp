// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package statemgr

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util"
)

// EventGetBlockMsg is a request for a block while syncing
func (sm *stateManager) EventGetBlockMsg(msg *chain.GetBlockMsg) {
	sm.eventGetBlockMsgCh <- msg
}
func (sm *stateManager) eventGetBlockMsg(msg *chain.GetBlockMsg) {
	if sm.stateOutput == nil || sm.solidState == nil {
		return
	}
	sm.log.Infow("WWW EventGetBlockMsg",
		"sender index", msg.SenderIndex,
		"block index", msg.BlockIndex,
	)
	sm.log.Debugw("EventGetBlockMsg",
		"sender index", msg.SenderIndex,
		"blockBytes index", msg.BlockIndex,
	)
	blockBytes, err := state.LoadBlockBytes(sm.dbp, sm.chain.ID(), msg.BlockIndex)
	if err != nil {
		return
	}

	sm.log.Infof("WWW EventGetBlockMsg for state index #%d --> peer %d", msg.BlockIndex, msg.SenderIndex)
	sm.log.Debugf("EventGetBlockMsg for state index #%d --> peer %d", msg.BlockIndex, msg.SenderIndex)

	err = sm.peers.SendMsg(msg.SenderIndex, chain.MsgBlock, util.MustBytes(&chain.BlockMsg{
		BlockBytes: blockBytes,
	}))
	if err != nil {
		sm.log.Infof("WWW EventGetBlockMsg: erro sending %v", err)
		return
	}
}

// EventBlockMsg
func (sm *stateManager) EventBlockMsg(msg *chain.BlockMsg) {
	sm.eventBlockMsgCh <- msg
}
func (sm *stateManager) eventBlockMsg(msg *chain.BlockMsg) {
	if sm.stateOutput == nil {
		return
	}
	block, err := state.BlockFromBytes(msg.BlockBytes)
	if err != nil {
		sm.log.Warnf("wrong blokc received from peer %d. Err: %v", msg.SenderIndex, err)
		return
	}
	sm.log.Infow("WWW EventBlockMsg",
		"sender", msg.SenderIndex,
		"block index", block.BlockIndex(),
		"approving output", coretypes.OID(block.ApprovingOutputID()),
	)
	sm.log.Debugw("EventBlockMsg",
		"sender", msg.SenderIndex,
		"block index", block.BlockIndex(),
		"approving output", coretypes.OID(block.ApprovingOutputID()),
	)
	sm.addBlockFromPeer(block)
	sm.takeAction()
}

func (sm *stateManager) EventOutputMsg(msg ledgerstate.Output) {
	sm.eventOutputMsgCh <- msg
}
func (sm *stateManager) eventOutputMsg(msg ledgerstate.Output) {
	sm.log.Debugf("EventOutputMsg: %s", coretypes.OID(msg.ID()))
	chainOutput, ok := msg.(*ledgerstate.AliasOutput)
	if !ok {
		return
	}
	sm.outputPulled(chainOutput)
	sm.takeAction()
}

// EventStateTransactionMsg triggered whenever new state transaction arrives
// the state transaction may be confirmed or not
func (sm *stateManager) EventStateMsg(msg *chain.StateMsg) {
	sm.eventStateOutputMsgCh <- msg
}
func (sm *stateManager) eventStateMsg(msg *chain.StateMsg) {
	stateHash, err := hashing.HashValueFromBytes(msg.ChainOutput.GetStateData())
	if err != nil {
		sm.log.Errorf("EventStateMsg: failed to parse state hash: %v", err)
		return
	}
	sm.log.Debugw("EventStateMsg",
		"state index", msg.ChainOutput.GetStateIndex(),
		"chainOutput", coretypes.OID(msg.ChainOutput.ID()),
		"state hash", stateHash.String(),
	)
	sm.outputPushed(msg.ChainOutput, msg.Timestamp)
	sm.takeAction()
}

func (sm *stateManager) EventStateCandidateMsg(msg chain.StateCandidateMsg) {
	sm.eventPendingBlockMsgCh <- msg
}
func (sm *stateManager) eventStateCandidateMsg(msg chain.StateCandidateMsg) {
	if sm.stateOutput == nil {
		return
	}
	sm.log.Debugf("EventStateCandidateMsg: state index: %d timestamp: %v",
		msg.State.BlockIndex(), msg.State.Timestamp(),
	)
	sm.addBlockFromCommitee(msg.State)
	sm.takeAction()
}

func (sm *stateManager) EventTimerMsg(msg chain.TimerTick) {
	if msg%2 == 0 {
		sm.eventTimerMsgCh <- msg
	}
}
func (sm *stateManager) eventTimerMsg(msg chain.TimerTick) {
	sm.takeAction()
}
