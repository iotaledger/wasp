package consensus

import (
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/coretypes"
)

func (c *consensus) EventStateTransitionMsg(msg *chain.StateTransitionMsg) {
	c.eventStateTransitionMsgCh <- msg
}
func (c *consensus) eventStateTransitionMsg(msg *chain.StateTransitionMsg) {
	c.log.Debugf("eventStateTransitionMsg: state index: %d, state output: %s, timestamp: %v",
		msg.State.BlockIndex(), coretypes.OID(msg.StateOutput.ID()), msg.StateTimestamp)
	c.setNewState(msg)
	c.takeAction()
}

func (c *consensus) EventSignedResultMsg(msg *chain.SignedResultMsg) {
	c.eventSignedResultMsgCh <- msg
}
func (c *consensus) eventSignedResult(msg *chain.SignedResultMsg) {
	c.log.Debugf("eventSignedResult: from sender: %d", msg.SenderIndex)
	c.receiveSignedResult(msg)
	c.takeAction()
}

func (c *consensus) EventInclusionsStateMsg(msg *chain.InclusionStateMsg) {
	c.eventInclusionStateMsgCh <- msg
}
func (c *consensus) eventInclusionState(msg *chain.InclusionStateMsg) {
	c.log.Debugf("eventInclusionState:  %s: '%s'", msg.TxID.Base58(), msg.State.String())
	c.processInclusionState(msg)

	c.takeAction()
}

func (c *consensus) EventAsynchronousCommonSubsetMsg(msg *chain.AsynchronousCommonSubsetMsg) {
	c.eventACSMsgCh <- msg
}
func (c *consensus) eventAsynchronousCommonSubset(msg *chain.AsynchronousCommonSubsetMsg) {
	c.log.Debugf("eventAsynchronousCommonSubset: len = %d", len(msg.ProposedBatchesBin))
	c.receiveACS(msg.ProposedBatchesBin, msg.SessionID)

	c.takeAction()
}

func (c *consensus) EventVMResultMsg(msg *chain.VMResultMsg) {
	c.eventVMResultMsgCh <- msg
}
func (c *consensus) eventVMResultMsg(msg *chain.VMResultMsg) {
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
