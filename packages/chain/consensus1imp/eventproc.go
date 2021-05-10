package consensus1imp

import (
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/coretypes"
)

func (c *consensusImpl) EventStateTransitionMsg(msg *chain.StateTransitionMsg) {
	c.eventStateTransitionMsgCh <- msg
}
func (c *consensusImpl) eventStateTransitionMsg(msg *chain.StateTransitionMsg) {
	c.log.Debugf("eventStateTransitionMsg: state index: %d, state output: %s, timestamp: %v",
		msg.State.BlockIndex(), coretypes.OID(msg.StateOutput.ID()), msg.StateTimestamp)
	c.stateOutput = msg.StateOutput
	c.currentState = msg.State
	c.stateTimestamp = msg.StateTimestamp

	c.takeAction()
}

func (c *consensusImpl) EventResultCalculated(msg *chain.VMResultMsg) {
	c.eventResultCalculatedMsgCh <- msg
}
func (c *consensusImpl) eventResultCalculated(msg *chain.VMResultMsg) {
	c.log.Debugf("eventResultCalculated: block index: %d", msg.Task.VirtualState.BlockIndex())

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
