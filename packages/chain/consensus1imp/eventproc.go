package consensus1imp

import "github.com/iotaledger/wasp/packages/chain"

func (c *consensusImpl) EventStateTransitionMsg(msg *chain.StateTransitionMsg) {
	c.eventStateTransitionMsgCh <- msg
}
func (c *consensusImpl) eventStateTransitionMsg(msg *chain.StateTransitionMsg) {

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
