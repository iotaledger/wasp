package consensus1imp

import "github.com/iotaledger/wasp/packages/chain"

func (c *consensusImpl) EventStateTransitionMsg(msg *chain.StateTransitionMsg) {
	panic("implement me")
}

func (c *consensusImpl) EventTimerMsg(tick chain.TimerTick) {
	panic("implement me")
}
