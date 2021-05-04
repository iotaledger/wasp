package chain

type Consensus1 interface {
	EventStateTransitionMsg(*StateTransitionMsg)
	EventTimerMsg(TimerTick)
}
