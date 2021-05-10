package chain

type Consensus1 interface {
	EventStateTransitionMsg(*StateTransitionMsg)
	EventResultCalculated(msg *VMResultMsg)
	EventSignedResultMsg(msg *SignedResultMsg)
	EventTimerMsg(TimerTick)
	IsReady() bool
	Close()
}
