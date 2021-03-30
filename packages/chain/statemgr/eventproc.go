// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package statemgr

import (
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util"
)

// EventPingPongMsg reacts to the PinPong message
func (sm *stateManager) EventStateIndexPingPongMsg(msg *chain.StateIndexPingPongMsg) {
	sm.eventStateIndexPingPongMsgCh <- msg
}
func (sm *stateManager) eventStateIndexPingPongMsg(msg *chain.StateIndexPingPongMsg) {
	before := sm.numPongsHasQuorum()
	sm.pingPongReceived(msg.SenderIndex)
	after := sm.numPongsHasQuorum()

	if msg.RSVP && sm.solidStateValid {
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
		PeerMsgHeader: chain.PeerMsgHeader{
			BlockIndex: msg.BlockIndex,
		},
		Size:                block.Size(),
		AnchorTransactionID: block.StateTransactionID(),
	}))
	if err != nil {
		return
	}
	block.ForEach(func(batchIndex uint16, stateUpdate state.StateUpdate) bool {
		err = sm.peers.SendMsg(msg.SenderIndex, chain.MsgStateUpdate, util.MustBytes(&chain.StateUpdateMsg{
			PeerMsgHeader: chain.PeerMsgHeader{
				BlockIndex: msg.BlockIndex,
			},
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
	sm.log.Debugw("EventBlockHeaderMsg",
		"sender", msg.SenderIndex,
		"state index", msg.BlockIndex,
		"size", msg.Size,
		"state tx", msg.AnchorTransactionID.String(),
	)
	if sm.syncedBatch != nil &&
		sm.syncedBatch.stateIndex == msg.BlockIndex &&
		sm.syncedBatch.stateTxId == msg.AnchorTransactionID &&
		len(sm.syncedBatch.stateUpdates) == int(msg.Size) {
		return // no need to start from scratch
	}
	sm.syncedBatch = &syncedBatch{
		stateIndex:   msg.BlockIndex,
		stateUpdates: make([]state.StateUpdate, msg.Size),
		stateTxId:    msg.AnchorTransactionID,
	}
}

// response to the state update msg.
// It collects state updates while waiting for the anchoring state transaction
func (sm *stateManager) EventStateUpdateMsg(msg *chain.StateUpdateMsg) {
	sm.eventStateUpdateMsgCh <- msg
}
func (sm *stateManager) eventStateUpdateMsg(msg *chain.StateUpdateMsg) {
	sm.log.Debugw("EventStateUpdateMsg",
		"sender", msg.SenderIndex,
		"state index", msg.BlockIndex,
		"block index", msg.IndexInTheBlock,
	)
	if sm.syncedBatch == nil {
		return
	}
	if sm.syncedBatch.stateIndex != msg.BlockIndex {
		return
	}
	if int(msg.IndexInTheBlock) >= len(sm.syncedBatch.stateUpdates) {
		sm.log.Errorf("bad block index in the state update message")
		return
	}
	sh := util.GetHashValue(msg.StateUpdate)
	sm.log.Debugf("EventStateUpdateMsg: receiving stateUpdate block index: %d hash: %s",
		msg.IndexInTheBlock, sh.String())

	sm.syncedBatch.stateUpdates[msg.IndexInTheBlock] = msg.StateUpdate
	sm.syncedBatch.msgCounter++

	if int(sm.syncedBatch.msgCounter) < len(sm.syncedBatch.stateUpdates) {
		// some are missing
		return
	}
	// check if whole block already received
	for _, su := range sm.syncedBatch.stateUpdates {
		if su == nil {
			// some state updates are missing
			return
		}
	}
	// the whole block received
	batch, err := state.NewBlock(sm.syncedBatch.stateUpdates...)
	if err != nil {
		sm.log.Errorf("failed to create block: %v", err)
		sm.syncedBatch = nil
		return
	}
	batch.WithBlockIndex(sm.syncedBatch.stateIndex).WithStateTransaction(sm.syncedBatch.stateTxId)

	sm.log.Debugf("EventStateUpdateMsg: reconstructed block %s", batch.String())

	sm.syncedBatch = nil
	go sm.chain.ReceiveMessage(chain.PendingBlockMsg{
		Block: batch,
	})
	sm.takeAction()
}

// EventStateTransactionMsg triggered whenever new state transaction arrives
// the state transaction may be confirmed or not
func (sm *stateManager) EventStateOutputMsg(msg *chain.StateMsg) {
	sm.eventStateOutputMsgCh <- msg
}
func (sm *stateManager) eventStateOutputMsg(msg *chain.StateMsg) {
	stateHash, err := hashing.HashValueFromBytes(msg.ChainOutput.GetStateData())
	if err != nil {
		sm.log.Panicf("failed to parse state hash: %v", err)
	}
	sm.log.Debugw("EventStateOutputMsg",
		"chainOutput", msg.ChainOutput.ID().Base58(),
		"state index", msg.ChainOutput.GetStateIndex(),
		"state hash", stateHash.String(),
	)

	sm.evidenceStateIndex(msg.ChainOutput.GetStateIndex())

	if sm.solidStateValid {
		if msg.ChainOutput.GetStateIndex() != sm.solidState.BlockIndex()+1 {
			sm.log.Debugf("skip state transaction: expected with state index #%d, got #%d, Txid: %s",
				sm.solidState.BlockIndex()+1, msg.ChainOutput.GetStateIndex(), msg.ChainOutput.ID().Base58())
			return
		}
	} else {
		if sm.solidState == nil {
			// pre-origin
			if msg.ChainOutput.GetStateIndex() != 0 {
				sm.log.Debugf("sm.solidState == nil && stateBlock.BlockIndex() != 0")
				return
			}
		} else {
			if msg.ChainOutput.GetStateIndex() != sm.solidState.BlockIndex() {
				sm.log.Debugf("sm.solidState == nil && stateBlock.BlockIndex() != sm.solidState.BlockIndex()")
				return
			}
		}
	}
	sm.nextStateOutput = msg.ChainOutput
	sm.nextStateOutputTimestamp = msg.Timestamp

	sm.takeAction()
}

func (sm *stateManager) EventPendingBlockMsg(msg chain.PendingBlockMsg) {
	sm.eventPendingBlockMsgCh <- msg
}
func (sm *stateManager) eventPendingBlockMsg(msg chain.PendingBlockMsg) {
	sm.log.Debugw("EventPendingBlockMsg",
		"state index", msg.Block.StateIndex(),
		"size", msg.Block.Size(),
		"txid", msg.Block.StateTransactionID().String(),
		"block essence", msg.Block.EssenceHash().String(),
		"ts", msg.Block.Timestamp(),
	)

	sm.addPendingBlock(msg.Block)
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
