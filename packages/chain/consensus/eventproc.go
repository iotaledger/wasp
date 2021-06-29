package consensus

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
)

func (c *Consensus) EventStateTransitionMsg(msg *chain.StateTransitionMsg) {
	c.eventStateTransitionMsgCh <- msg
}

func (c *Consensus) eventStateTransitionMsg(msg *chain.StateTransitionMsg) {
	c.log.Debugf("StateTransitionMsg received: state index: %d, state output: %s, timestamp: %v",
		msg.State.BlockIndex(), coretypes.OID(msg.StateOutput.ID()), msg.StateTimestamp)
	c.setNewState(msg)
	c.takeAction()
}

func (c *Consensus) EventSignedResultMsg(msg *chain.SignedResultMsg) {
	c.eventSignedResultMsgCh <- msg
}

func (c *Consensus) eventSignedResult(msg *chain.SignedResultMsg) {
	c.log.Debugf("SignedResultMsg received: from sender %d, hash=%s, chain input id=%v",
		msg.SenderIndex, msg.EssenceHash, coretypes.OID(msg.ChainInputID))
	c.receiveSignedResult(msg)
	c.takeAction()
}

func (c *Consensus) EventInclusionsStateMsg(msg *chain.InclusionStateMsg) {
	c.eventInclusionStateMsgCh <- msg
}

func (c *Consensus) eventInclusionState(msg *chain.InclusionStateMsg) {
	c.log.Debugf("InclusionStateMsg received:  %s: '%s'", msg.TxID.Base58(), msg.State.String())
	c.processInclusionState(msg)

	c.takeAction()
}

func (c *Consensus) EventAsynchronousCommonSubsetMsg(msg *chain.AsynchronousCommonSubsetMsg) {
	c.eventACSMsgCh <- msg
}

func (c *Consensus) eventAsynchronousCommonSubset(msg *chain.AsynchronousCommonSubsetMsg) {
	c.log.Debugf("AsynchronousCommonSubsetMsg received for session %v: len = %d", msg.SessionID, len(msg.ProposedBatchesBin))
	c.receiveACS(msg.ProposedBatchesBin, msg.SessionID)

	c.takeAction()
}

func (c *Consensus) EventVMResultMsg(msg *chain.VMResultMsg) {
	c.eventVMResultMsgCh <- msg
}

func (c *Consensus) eventVMResultMsg(msg *chain.VMResultMsg) {
	var essenceString string
	if msg.Task.ResultTransactionEssence == nil {
		essenceString = "essence is nil"
	} else {
		essenceString = fmt.Sprintf("essence hash: %s", hashing.HashData(msg.Task.ResultTransactionEssence.Bytes()))
	}
	c.log.Debugf("VMResultMsg received: state index: %d state hash: %s %s",
		msg.Task.VirtualState.BlockIndex(), msg.Task.VirtualState.Hash(), essenceString)
	c.processVMResult(msg.Task)
	c.takeAction()
}

func (c *Consensus) EventTimerMsg(msg chain.TimerTick) {
	c.eventTimerMsgCh <- msg
}

func (c *Consensus) eventTimerMsg(msg chain.TimerTick) {
	c.lastTimerTick.Store(int64(msg))
	c.refreshConsensusInfo()
	if msg%40 == 0 {
		if snap := c.GetStatusSnapshot(); snap != nil {
			c.log.Infof("timer tick #%d: state index: %d, mempool = (%d, %d)",
				snap.TimerTick, snap.StateIndex, snap.Mempool.InPoolCounter, snap.Mempool.OutPoolCounter)
		}
	}
	c.takeAction()
}
