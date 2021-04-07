// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package statemgr

import (
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util"
)

// EventPingPongMsg reacts to the PinPong message
func (sm *stateManager) EventStateIndexPingPongMsg(msg *chain.BlockIndexPingPongMsg) {
	sm.eventStateIndexPingPongMsgCh <- msg
}
func (sm *stateManager) eventStateIndexPingPongMsg(msg *chain.BlockIndexPingPongMsg) {
	before := sm.numPongsHasQuorum()
	sm.pingPongReceived(msg.SenderIndex)
	after := sm.numPongsHasQuorum()

	if msg.RSVP && sm.isSolidStateValidated() {
		// only respond if current solid state is validated
		sm.respondPongToPeer(msg.SenderIndex)
	}
	if !before && after {
		sm.log.Infof("collected %d evidences of state indices of peers", sm.numPongs())
	}
	//sm.log.Debugw("EventStateIndexPingPongMsg",
	//	"sender", msg.SenderIndex,
	//	"state", msg.BlockIndex,
	//	"rsvp", msg.RSVP,
	//	"numPongs", sm.numPongs(),
	//)
}

// EventGetBlockMsg is a request for a block while syncing
func (sm *stateManager) EventGetBlockMsg(msg *chain.GetBlockMsg) {
	sm.eventGetBlockMsgCh <- msg
}
func (sm *stateManager) eventGetBlockMsg(msg *chain.GetBlockMsg) {
	if sm.stateOutput == nil {
		return
	}
	sm.log.Debugw("EventGetBlockMsg",
		"sender index", msg.SenderIndex,
		"block index", msg.BlockIndex,
	)
	block, err := state.LoadBlock(sm.chain.ID(), msg.BlockIndex)
	if err != nil || block == nil {
		// can't load block, can't respond
		return
	}

	sm.log.Debugf("EventGetBlockMsg for state index #%d --> peer %d", msg.BlockIndex, msg.SenderIndex)

	err = sm.peers.SendMsg(msg.SenderIndex, chain.MsgBatchHeader, util.MustBytes(&chain.BlockHeaderMsg{
		BlockIndex:        msg.BlockIndex,
		Size:              block.Size(),
		ApprovingOutputID: block.ApprovingOutputID(),
	}))
	if err != nil {
		return
	}
	block.ForEach(func(batchIndex uint16, stateUpdate state.StateUpdate) bool {
		err = sm.peers.SendMsg(msg.SenderIndex, chain.MsgStateUpdate, util.MustBytes(&chain.StateUpdateMsg{
			BlockIndex:      msg.BlockIndex,
			StateUpdate:     stateUpdate,
			IndexInTheBlock: batchIndex,
		}))
		sh := util.GetHashValue(stateUpdate)
		sm.log.Debugw("EventGetBlockMsg: sending stateUpdate", "hash", sh.String())
		return true
	})
}

// EventBlockHeaderMsg
func (sm *stateManager) EventBlockHeaderMsg(msg *chain.BlockHeaderMsg) {
	sm.eventBlockHeaderMsgCh <- msg
}
func (sm *stateManager) eventBlockHeaderMsg(msg *chain.BlockHeaderMsg) {
	if sm.stateOutput == nil {
		return
	}
	sm.log.Debugw("EventBlockHeaderMsg",
		"sender", msg.SenderIndex,
		"state index", msg.BlockIndex,
		"size", msg.Size,
		"state tx", msg.ApprovingOutputID.String(),
	)
	sm.blockHeaderArrived(msg)
	sm.takeAction()
}

// response to the state update msg.
// It collects state updates while waiting for the anchoring state transaction
func (sm *stateManager) EventStateUpdateMsg(msg *chain.StateUpdateMsg) {
	sm.eventStateUpdateMsgCh <- msg
}
func (sm *stateManager) eventStateUpdateMsg(msg *chain.StateUpdateMsg) {
	if sm.stateOutput == nil {
		return
	}
	sm.log.Debugw("EventStateUpdateMsg",
		"sender", msg.SenderIndex,
		"state index", msg.BlockIndex,
		"block index", msg.IndexInTheBlock,
	)
	sm.stateUpdateArrived(msg)
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
	if sm.stateOutput != nil {
		switch {
		case sm.stateOutput.GetStateIndex() == msg.ChainOutput.GetStateIndex():
			sm.log.Debug("EventStateMsg: repeated state output")
			return
		case sm.stateOutput.GetStateIndex() > msg.ChainOutput.GetStateIndex():
			sm.log.Warn("EventStateMsg: out of order state output")
			return
		}
	}
	sm.stateOutput = msg.ChainOutput
	sm.stateOutputTimestamp = msg.Timestamp
	sm.checkStateApproval()
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

	sm.addBlockCandidate(msg.Block)
	sm.checkStateApproval()
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
