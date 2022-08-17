// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package consensus

import (
	"fmt"
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/chain/messages"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/state"
)

func (c *consensus) EnqueueStateTransitionMsg(isGovernance bool, virtualState state.VirtualStateAccess, stateOutput *isc.AliasOutputWithID, stateTimestamp time.Time) {
	c.eventStateTransitionMsgPipe.In() <- &messages.StateTransitionMsg{
		IsGovernance:   isGovernance,
		State:          virtualState,
		StateOutput:    stateOutput,
		StateTimestamp: stateTimestamp,
	}
}

func (c *consensus) handleStateTransitionMsg(msg *messages.StateTransitionMsg) {
	c.log.Debugf("StateTransitionMsg received: governance updated: %v, state index: %d, state output: %s, timestamp: %v",
		msg.IsGovernance, msg.State.BlockIndex(), isc.OID(msg.StateOutput.ID()), msg.StateTimestamp)
	if c.setNewState(msg) {
		c.takeAction()
	}
}

func (c *consensus) EnqueueDssIndexProposalMsg(msg *messages.DssIndexProposalMsg) {
	c.eventDssIndexProposalMsgPipe.In() <- msg
}

func (c *consensus) handleDssIndexProposalMsg(msg *messages.DssIndexProposalMsg) {
	c.log.Debugf("handleDssIndexProposalMsg message received for %s: index proposal: %v", msg.DssKey, msg.IndexProposal)
	c.receiveDssIndexProposal(msg.DssKey, msg.IndexProposal)
	c.takeAction()
}

func (c *consensus) EnqueueDssSignatureMsg(msg *messages.DssSignatureMsg) {
	c.eventDssSignatureMsgPipe.In() <- msg
}

func (c *consensus) handleDssSignatureMsg(msg *messages.DssSignatureMsg) {
	c.log.Debugf("handleDssSignatureMsg message received for %s: signature: %v", msg.DssKey, msg.Signature)
	c.receiveDssSignature(msg.DssKey, msg.Signature)
	c.takeAction()
}

// func (c *consensus) EnqueueSignedResultMsg(msg *messages.SignedResultMsgIn) {
// 	c.eventSignedResultMsgPipe.In() <- msg
// }

// func (c *consensus) handleSignedResultMsg(msg *messages.SignedResultMsgIn) {
// 	c.log.Debugf("handleSignedResultMsg message received: from sender %d, hash=%s, chain input id=%v",
// 		msg.SenderIndex, msg.EssenceHash, isc.OID(msg.ChainInputID))
// 	c.receiveSignedResult(msg)
// 	c.takeAction()
// }

// func (c *consensus) EnqueueSignedResultAckMsg(msg *messages.SignedResultAckMsgIn) {
// 	c.eventSignedResultAckMsgPipe.In() <- msg
// }

// func (c *consensus) handleSignedResultAckMsg(msg *messages.SignedResultAckMsgIn) {
// 	c.log.Debugf("SignedResultAckMsg received: from sender %d, hash=%s, chain input id=%v",
// 		msg.SenderIndex, msg.EssenceHash, isc.OID(msg.ChainInputID))
// 	c.receiveSignedResultAck(msg)
// 	c.takeAction()
// }

func (c *consensus) EnqueueTxInclusionsStateMsg(txID iotago.TransactionID, inclusionState string) {
	c.eventInclusionStateMsgPipe.In() <- &messages.TxInclusionStateMsg{
		TxID:  txID,
		State: inclusionState,
	}
}

func (c *consensus) handleTxInclusionState(msg *messages.TxInclusionStateMsg) {
	c.log.Debugf("TxInclusionStateMsg received:  %s: '%s'", isc.TxID(msg.TxID), msg.State)
	c.processTxInclusionState(msg)

	c.takeAction()
}

func (c *consensus) EnqueueAsynchronousCommonSubsetMsg(msg *messages.AsynchronousCommonSubsetMsg) {
	c.eventACSMsgPipe.In() <- msg
}

func (c *consensus) handleAsynchronousCommonSubset(msg *messages.AsynchronousCommonSubsetMsg) {
	c.log.Debugf("AsynchronousCommonSubsetMsg received for session %v: len = %d", msg.SessionID, len(msg.ProposedBatchesBin))
	c.receiveACS(msg.ProposedBatchesBin, msg.SessionID)

	c.takeAction()
}

func (c *consensus) EnqueueVMResultMsg(msg *messages.VMResultMsg) {
	c.eventVMResultMsgPipe.In() <- msg
}

func (c *consensus) handleVMResultMsg(msg *messages.VMResultMsg) {
	var essenceString string
	if msg.Task.ResultTransactionEssence == nil {
		essenceString = "essence is nil"
	} else {
		signingMsg, err := msg.Task.ResultTransactionEssence.SigningMessage()
		if err != nil {
			essenceString = fmt.Sprintf("essence signing message not retrievable: %v", err)
		} else {
			essenceString = fmt.Sprintf("essence signing message hash: %s", hashing.HashData(signingMsg))
		}
	}
	c.log.Debugf("VMResultMsg received: state index: %d state commitment: %s %s",
		msg.Task.VirtualStateAccess.BlockIndex(), state.RootCommitment(msg.Task.VirtualStateAccess.TrieNodeStore()), essenceString)
	c.processVMResult(msg.Task)
	c.takeAction()
}

func (c *consensus) EnqueueTimerMsg(msg messages.TimerTick) {
	c.eventTimerMsgPipe.In() <- msg
}

func (c *consensus) handleTimerMsg(msg messages.TimerTick) {
	c.lastTimerTick.Store(int64(msg))
	c.refreshConsensusInfo()
	if msg%40 == 0 {
		if snap := c.GetStatusSnapshot(); snap != nil {
			c.log.Infof("timer tick #%d: state index: %d, mempool = (total: %d, ready: %d, in: %d, out: %d)",
				snap.TimerTick, snap.StateIndex, snap.Mempool.TotalPool, snap.Mempool.ReadyCounter, snap.Mempool.InPoolCounter, snap.Mempool.OutPoolCounter)
		}
	}
	c.takeAction()
	if c.stateOutput != nil {
		c.log.Debugf("Consensus::eventTimerMsg: stateIndex=%v, workflow=%+v",
			c.stateOutput.GetStateIndex(),
			c.workflow,
		)
	} else {
		c.log.Debugf("Consensus::eventTimerMsg: stateIndex=nil, workflow=%+v",
			c.workflow,
		)
	}
}
