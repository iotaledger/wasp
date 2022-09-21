package consensus

type pipeMetrics struct {
	eventStateTransitionMsgPipeSize int
	eventPeerLogIndexMsgPipeSize    int
	eventInclusionStateMsgPipeSize  int
	eventACSMsgPipeSize             int
	eventVMResultMsgPipeSize        int
	eventTimerMsgPipeSize           int
}

func (p *pipeMetrics) GetEventStateTransitionMsgPipeSize() int {
	return p.eventStateTransitionMsgPipeSize
}

func (p *pipeMetrics) GetEventPeerLogIndexMsgPipeSize() int {
	return p.eventPeerLogIndexMsgPipeSize
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
