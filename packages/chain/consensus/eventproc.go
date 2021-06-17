package consensus

import (
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
)

func (c *consensus) EventStateTransitionMsg(msg *chain.StateTransitionMsg) {
	c.eventStateTransitionMsgCh <- msg
}
func (c *consensus) eventStateTransitionMsg(msg *chain.StateTransitionMsg) {
	c.log.Debugf("StateTransitionMsg received: state index: %d, state output: %s, timestamp: %v",
		msg.State.BlockIndex(), coretypes.OID(msg.StateOutput.ID()), msg.StateTimestamp)
	c.setNewState(msg)
	c.takeAction()
}

func (c *consensus) EventSignedResultMsg(msg *chain.SignedResultMsg) {
	c.eventSignedResultMsgCh <- msg
}
func (c *consensus) eventSignedResult(msg *chain.SignedResultMsg) {
	c.log.Debugf("SignedResultMsg received: from sender %d, hash=%s, chain input id=%v",
		msg.SenderIndex, msg.EssenceHash, coretypes.OID(msg.ChainInputID))
	c.receiveSignedResult(msg)
	c.takeAction()
}

func (c *consensus) EventInclusionsStateMsg(msg *chain.InclusionStateMsg) {
	c.eventInclusionStateMsgCh <- msg
}
func (c *consensus) eventInclusionState(msg *chain.InclusionStateMsg) {
	c.log.Debugf("InclusionStateMsg received:  %s: '%s'", msg.TxID.Base58(), msg.State.String())
	c.processInclusionState(msg)

	c.takeAction()
}

func (c *consensus) EventAsynchronousCommonSubsetMsg(msg *chain.AsynchronousCommonSubsetMsg) {
	c.eventACSMsgCh <- msg
}
func (c *consensus) eventAsynchronousCommonSubset(msg *chain.AsynchronousCommonSubsetMsg) {
	c.log.Debugf("AsynchronousCommonSubsetMsg received for session %v: len = %d", msg.SessionID, len(msg.ProposedBatchesBin))
	c.receiveACS(msg.ProposedBatchesBin, msg.SessionID)

	c.takeAction()
}

func (c *consensus) EventVMResultMsg(msg *chain.VMResultMsg) {
	c.eventVMResultMsgCh <- msg
}
func (c *consensus) eventVMResultMsg(msg *chain.VMResultMsg) {
	essenceHash := hashing.HashData(msg.Task.ResultTransactionEssence.Bytes())
	c.log.Debugf("VMResultMsg received: state index: %d state hash: %s essence hash: %s",
		msg.Task.VirtualState.BlockIndex(), msg.Task.VirtualState.Hash(), essenceHash)
	c.processVMResult(msg.Task)
	c.takeAction()
}

func (c *consensus) EventTimerMsg(msg chain.TimerTick) {
	c.eventTimerMsgCh <- msg
}
func (c *consensus) eventTimerMsg(msg chain.TimerTick) {
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
