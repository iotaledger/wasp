package cmt_log

import "github.com/iotaledger/wasp/packages/gpa"

// This event is introduced to avoid too-fast recovery from the
// L1 rejections, because L1 sometimes report them prematurely.
type inputMilestoneReceived struct{}

func NewInputMilestoneReceived() gpa.Input {
	return &inputMilestoneReceived{}
}

func (inp *inputMilestoneReceived) String() string {
	return "{cmtLog.inputMilestoneReceived}"
}
