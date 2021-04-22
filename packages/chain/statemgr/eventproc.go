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
	sm.log.Infow("XXX EventGetBlockMsg",
		"sender index", msg.SenderIndex,
		"block index", msg.BlockIndex,
	)
	sm.log.Debugw("EventGetBlockMsg",
		"sender index", msg.SenderIndex,
		"block index", msg.BlockIndex,
	)
	block, err := state.LoadBlock(sm.dbp.GetPartition(sm.chain.ID()), msg.BlockIndex)
	if err != nil || block == nil {
		sm.log.Infof("XXX EventGetBlockMsg: can't load message %v: %v", msg.BlockIndex, err)
		// can't load block, can't respond
		return
	}

	sm.log.Infof("XXX EventGetBlockMsg for state index #%d --> peer %d", msg.BlockIndex, msg.SenderIndex)
	sm.log.Debugf("EventGetBlockMsg for state index #%d --> peer %d", msg.BlockIndex, msg.SenderIndex)

	err = sm.peers.SendMsg(msg.SenderIndex, chain.MsgBlock, util.MustBytes(&chain.BlockMsg{
		Block: block,
	}))
	if err != nil {
		sm.log.Infof("XXX EventGetBlockMsg: erro sending %v", err)
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
	sm.log.Infow("XXX EventBlockMsg",
		"sender", msg.SenderIndex,
		"block index", msg.Block.StateIndex(),
		"essence hash", msg.Block.EssenceHash().String(),
		"approving output", coretypes.OID(msg.Block.ApprovingOutputID()),
	)
	sm.log.Debugw("EventBlockMsg",
		"sender", msg.SenderIndex,
		"block index", msg.Block.StateIndex(),
		"essence hash", msg.Block.EssenceHash().String(),
		"approving output", coretypes.OID(msg.Block.ApprovingOutputID()),
	)
	sm.addBlockFromNode(msg.Block)
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

func (sm *stateManager) EventBlockCandidateMsg(msg chain.BlockCandidateMsg) {
	sm.eventPendingBlockMsgCh <- msg
}
func (sm *stateManager) eventBlockCandidateMsg(msg chain.BlockCandidateMsg) {
	if sm.stateOutput == nil {
		return
	}
	sm.log.Debugw("EventBlockCandidateMsg",
		"state index", msg.Block.StateIndex(),
		"size", msg.Block.Size(),
		"state output", coretypes.OID(msg.Block.ApprovingOutputID()),
		"block essence", msg.Block.EssenceHash().String(),
		"ts", msg.Block.Timestamp(),
	)
	sm.addBlockFromCommitee(msg.Block)
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
