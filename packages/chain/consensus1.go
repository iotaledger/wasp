package chain

type Consensus1 interface {
	EventStateTransitionMsg(*StateTransitionMsg)
	EventResultCalculated(msg *VMResultMsg)
	EventTimerMsg(TimerTick)
	IsReady() bool
	Close()
}
