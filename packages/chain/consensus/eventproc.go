// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package consensus

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/chain/messages"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
)

func (c *Consensus) EventStateTransitionMsg(msg *messages.StateTransitionMsg) {
	c.eventStateTransitionMsgCh <- msg
}

func (c *Consensus) eventStateTransitionMsg(msg *messages.StateTransitionMsg) {
	c.log.Debugf("StateTransitionMsg received: state index: %d, state output: %s, timestamp: %v",
		msg.State.BlockIndex(), iscp.OID(msg.StateOutput.ID()), msg.StateTimestamp)
	c.setNewState(msg)
	c.takeAction()
}

func (c *Consensus) EventSignedResultMsg(msg *messages.SignedResultMsg) {
	c.eventSignedResultMsgCh <- msg
}

func (c *Consensus) eventSignedResult(msg *messages.SignedResultMsg) {
	c.log.Debugf("SignedResultMsg received: from sender %d, hash=%s, chain input id=%v",
		msg.SenderIndex, msg.EssenceHash, iscp.OID(msg.ChainInputID))
	c.receiveSignedResult(msg)
	c.takeAction()
}

func (c *Consensus) EventSignedResultAckMsg(msg *messages.SignedResultAckMsg) {
	c.eventSignedResultAckMsgCh <- msg
}

func (c *Consensus) eventSignedResultAck(msg *messages.SignedResultAckMsg) {
	c.log.Debugf("SignedResultAckMsg received: from sender %d, hash=%s, chain input id=%v",
		msg.SenderIndex, msg.EssenceHash, iscp.OID(msg.ChainInputID))
	c.receiveSignedResultAck(msg)
	c.takeAction()
}

func (c *Consensus) EventInclusionsStateMsg(msg *messages.InclusionStateMsg) {
	c.eventInclusionStateMsgCh <- msg
}

func (c *Consensus) eventInclusionState(msg *messages.InclusionStateMsg) {
	c.log.Debugf("InclusionStateMsg received:  %s: '%s'", msg.TxID.Base58(), msg.State.String())
	c.processInclusionState(msg)

	c.takeAction()
}

func (c *Consensus) EventAsynchronousCommonSubsetMsg(msg *messages.AsynchronousCommonSubsetMsg) {
	c.eventACSMsgCh <- msg
}

func (c *Consensus) eventAsynchronousCommonSubset(msg *messages.AsynchronousCommonSubsetMsg) {
	c.log.Debugf("AsynchronousCommonSubsetMsg received for session %v: len = %d", msg.SessionID, len(msg.ProposedBatchesBin))
	c.receiveACS(msg.ProposedBatchesBin, msg.SessionID)

	c.takeAction()
}

func (c *Consensus) EventVMResultMsg(msg *messages.VMResultMsg) {
	c.eventVMResultMsgCh <- msg
}

func (c *Consensus) eventVMResultMsg(msg *messages.VMResultMsg) {
	var essenceString string
	if msg.Task.ResultTransactionEssence == nil {
		essenceString = "essence is nil"
	} else {
		essenceString = fmt.Sprintf("essence hash: %s", hashing.HashData(msg.Task.ResultTransactionEssence.Bytes()))
	}
	c.log.Debugf("VMResultMsg received: state index: %d state hash: %s %s",
		msg.Task.VirtualState.BlockIndex(), msg.Task.VirtualState.StateCommitment(), essenceString)
	c.processVMResult(msg.Task)
	c.takeAction()
}

func (c *Consensus) EventTimerMsg(msg messages.TimerTick) {
	c.eventTimerMsgCh <- msg
}

func (c *Consensus) eventTimerMsg(msg messages.TimerTick) {
	c.lastTimerTick.Store(int64(msg))
	c.refreshConsensusInfo()
	if msg%40 == 0 {
		if snap := c.GetStatusSnapshot(); snap != nil {
			c.log.Infof("timer tick #%d: state index: %d, mempool = (%d, %d, %d)",
				snap.TimerTick, snap.StateIndex, snap.Mempool.InPoolCounter, snap.Mempool.OutPoolCounter, snap.Mempool.TotalPool)
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
