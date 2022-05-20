package consensus

type pipeMetrics struct {
	eventStateTransitionMsgPipeSize int
	eventSignedResultMsgPipeSize    int
	eventSignedResultAckMsgPipeSize int
	eventInclusionStateMsgPipeSize  int
	eventACSMsgPipeSize             int
	eventVMResultMsgPipeSize        int
	eventTimerMsgPipeSize           int
}

func (p *pipeMetrics) GetEventStateTransitionMsgPipeSize() int {
	return p.eventStateTransitionMsgPipeSize
}

func (p *pipeMetrics) GetEventSignedResultMsgPipeSize() int {
	return p.eventSignedResultMsgPipeSize
}

func (p *pipeMetrics) GetEventSignedResultAckMsgPipeSize() int {
	return p.eventSignedResultAckMsgPipeSize
}

func (p *pipeMetrics) GetEventInclusionStateMsgPipeSize() int {
	return p.eventInclusionStateMsgPipeSize
}

func (p *pipeMetrics) GetEventACSMsgPipeSize() int {
	return p.eventACSMsgPipeSize
}

func (p *pipeMetrics) GetEventVMResultMsgPipeSize() int {
	return p.eventVMResultMsgPipeSize
}

func (p *pipeMetrics) GetEventTimerMsgPipeSize() int {
	return p.eventTimerMsgPipeSize
}
