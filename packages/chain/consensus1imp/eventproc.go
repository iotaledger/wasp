package consensus1imp

import (
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/coretypes"
)

func (c *consensusImpl) EventStateTransitionMsg(msg *chain.StateTransitionMsg) {
	c.eventStateTransitionMsgCh <- msg
}
func (c *consensusImpl) eventStateTransitionMsg(msg *chain.StateTransitionMsg) {
	c.log.Debugf("eventStateTransitionMsg: stage: '%s', state index: %d, state output: %s, timestamp: %v",
		c.stage, msg.State.BlockIndex(), coretypes.OID(msg.StateOutput.ID()), msg.StateTimestamp)
	c.setNewState(msg)
	c.takeAction()
}

func (c *consensusImpl) EventResultCalculated(msg *chain.VMResultMsg) {
	c.eventResultCalculatedMsgCh <- msg
}
func (c *consensusImpl) eventResultCalculated(msg *chain.VMResultMsg) {
	c.log.Debugf("eventResultCalculated: stage: '%s', block index: %d", c.stage, msg.Task.VirtualState.BlockIndex())

	if c.stage != stageVM || msg.Task.ChainInput.ID() != c.stateOutput.ID() {
		c.log.Warnf("eventResultCalculated: VMResultMsg out of context")
		return
	}
	c.processVMResult(msg.Task)
	c.takeAction()
}

func (c *consensusImpl) EventSignedResultMsg(msg *chain.SignedResultMsg) {
	c.eventSignedResultMsgCh <- msg
}
func (c *consensusImpl) eventSignedResult(msg *chain.SignedResultMsg) {
	c.log.Debugf("eventSignedResult: stage: '%s', from sender: %d", c.stage, msg.SenderIndex)
	c.processSignedResult(msg)
	c.takeAction()
}

func (c *consensusImpl) EventInclusionsStateMsg(msg *chain.InclusionStateMsg) {
	c.eventInclusionStateMsgCh <- msg
}
func (c *consensusImpl) eventInclusionState(msg *chain.InclusionStateMsg) {
	c.log.Debugf("eventInclusionState: stage: '%s',  %s: '%s'", c.stage, msg.TxID.Base58(), msg.State.String())

	c.takeAction()
}

func (c *consensusImpl) EventAsynchronousCommonSubsetMsg(msg *chain.AsynchronousCommonSubsetMsg) {
	c.eventACSMsgCh <- msg
}
func (c *consensusImpl) eventAsynchronousCommonSubset(msg *chain.AsynchronousCommonSubsetMsg) {
	c.log.Debugf("eventAsynchronousCommonSubset: stage: '%s' len = %d", c.stage, len(msg.ProposedBatchesBin))
	c.receiveACS(msg.ProposedBatchesBin)

	c.takeAction()
}

func (c *consensusImpl) EventVMResultMsg(msg *chain.VMResultMsg) {
	c.eventVMResultMsgCh <- msg
}
func (c *consensusImpl) eventVMResultMsg(msg *chain.VMResultMsg) {
	c.log.Debugf("eventVMResultMsg: stage: '%s' state index: %d state hash: %s",
		c.stage, msg.Task.VirtualState.BlockIndex(), msg.Task.VirtualState.Hash())
	c.processVMResult(msg.Task)

	c.takeAction()
}

func (c *consensusImpl) EventTimerMsg(msg chain.TimerTick) {
	c.eventTimerMsgCh <- msg
}
func (c *consensusImpl) eventTimerMsg(msg chain.TimerTick) {
	if msg%40 == 0 {
		c.log.Infof("timer tick #%d", msg)
	}
	c.lastTimerTick.Store(int64(msg))
	c.takeAction()
}

// for testing
func (c *consensusImpl) getTimerTick() int {
	return int(c.lastTimerTick.Load())
}

func (c *consensusImpl) getStateIndex() uint32 {
	return c.stateIndex.Load()
}
