package consensus1imp

import (
	"time"

	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/coretypes"
)

func (c *consensusImpl) EventStateTransitionMsg(msg *chain.StateTransitionMsg) {
	c.eventStateTransitionMsgCh <- msg
}
func (c *consensusImpl) eventStateTransitionMsg(msg *chain.StateTransitionMsg) {
	c.log.Debugf("eventStateTransitionMsg: state index: %d, state output: %s, timestamp: %v",
		msg.State.BlockIndex(), coretypes.OID(msg.StateOutput.ID()), msg.StateTimestamp)
	c.setNewState(msg)
	c.takeAction()
}

func (c *consensusImpl) EventResultCalculated(msg *chain.VMResultMsg) {
	c.eventResultCalculatedMsgCh <- msg
}
func (c *consensusImpl) eventResultCalculated(msg *chain.VMResultMsg) {
	c.log.Debugf("eventResultCalculated: block index: %d", msg.Task.VirtualState.BlockIndex())

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
	c.log.Debugf("eventSignedResult: from sender: %d", msg.SenderIndex)
	c.processSignedResult(msg)
	c.takeAction()
}

func (c *consensusImpl) EventNotifyTransactionPosted(msg *chain.NotifyFinalResultPostedMsg) {
	c.eventNotifyTxPostedMsgCh <- msg
}
func (c *consensusImpl) eventNotifyTransactionPosted(msg *chain.NotifyFinalResultPostedMsg) {
	c.log.Debugf("eventNotifyTransactionPosted: from sender: %d", msg.SenderIndex)

	if c.stage == stageTransactionFinalized {
		c.stage = stageTransactionPosted
		c.stageStarted = time.Now()
	}
	// TODO query inclusion state
	c.takeAction()
}

func (c *consensusImpl) EventTimerMsg(msg chain.TimerTick) {
	c.eventTimerMsgCh <- msg
}
func (c *consensusImpl) eventTimerMsg(msg chain.TimerTick) {
	if msg%40 == 0 {
		c.log.Infow("timer tick",
			"#", msg,
		)
	}
	c.takeAction()
}
