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
	sm.log.Debugw("EventGetBlockMsg received: ",
		"sender", msg.SenderNetID,
		"blockBytes index", msg.BlockIndex,
	)
	if sm.stateOutput == nil {
		sm.log.Debugf("EventGetBlockMsg ignored: stateOutput is nil")
		return
	}
	blockBytes, err := state.LoadBlockBytes(sm.dbp, sm.chain.ID(), msg.BlockIndex)
	if blockBytes == nil {
		sm.log.Debugf("EventGetBlockMsg ignored: block not found")
		return
	}
	if err != nil {
		sm.log.Debugf("EventGetBlockMsg ignored: LoadBlockBytes error %v", err)
		return
	}

	sm.log.Debugf("EventGetBlockMsg for state index #%d --> responding to peer %s", msg.BlockIndex, msg.SenderNetID)

	sm.peers.SendSimple(msg.SenderNetID, chain.MsgBlock, util.MustBytes(&chain.BlockMsg{
		BlockBytes: blockBytes,
	}))
}

// EventBlockMsg
func (sm *stateManager) EventBlockMsg(msg *chain.BlockMsg) {
	sm.eventBlockMsgCh <- msg
}
func (sm *stateManager) eventBlockMsg(msg *chain.BlockMsg) {
	sm.log.Debugf("EventBlockMsg received from %v", msg.SenderNetID)
	if sm.stateOutput == nil {
		sm.log.Debugf("EventBlockMsg ignored: stateOutput is nil")
		return
	}
	block, err := state.BlockFromBytes(msg.BlockBytes)
	if err != nil {
		sm.log.Warnf("EventBlockMsg ignored: wrong block received from peer %s. Err: %v", msg.SenderNetID, err)
		return
	}
	sm.log.Debugw("EventBlockMsg from ",
		"sender", msg.SenderNetID,
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
	sm.log.Debugf("EventOutputMsg received: %s", coretypes.OID(msg.ID()))
	chainOutput, ok := msg.(*ledgerstate.AliasOutput)
	if !ok {
		sm.log.Debugf("EventOutputMsg ignored: output is of type %t, expecting *ledgerstate.AliasOutput", msg)
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
	sm.log.Debugw("EventStateMsg received: ",
		"state index", msg.ChainOutput.GetStateIndex(),
		"chainOutput", coretypes.OID(msg.ChainOutput.ID()),
	)
	stateHash, err := hashing.HashValueFromBytes(msg.ChainOutput.GetStateData())
	if err != nil {
		sm.log.Errorf("EventStateMsg ignored: failed to parse state hash: %v", err)
		return
	}
	sm.log.Debugf("EventStateMsg state hash is %v", stateHash.String())
	sm.outputPushed(msg.ChainOutput, msg.Timestamp)
	sm.takeAction()
}

func (sm *stateManager) EventStateCandidateMsg(msg chain.StateCandidateMsg) {
	sm.eventPendingBlockMsgCh <- msg
}
func (sm *stateManager) eventStateCandidateMsg(msg chain.StateCandidateMsg) {
	sm.log.Debugf("EventStateCandidateMsg received: state index: %d, timestamp: %v",
		msg.State.BlockIndex(), msg.State.Timestamp(),
	)
	if sm.stateOutput == nil {
		sm.log.Debugf("EventStateCandidateMsg ignored: stateOutput is nil")
		return
	}
	sm.addBlockFromConsensus(msg.State)
	sm.takeAction()
}

func (sm *stateManager) EventTimerMsg(msg chain.TimerTick) {
	if msg%2 == 0 {
		sm.eventTimerMsgCh <- msg
	}
}
func (sm *stateManager) eventTimerMsg(msg chain.TimerTick) {
	sm.log.Debugf("EventTimerMsg received")
	sm.takeAction()
}
