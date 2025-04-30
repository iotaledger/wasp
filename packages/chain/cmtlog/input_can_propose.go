package cmtlog

import "github.com/iotaledger/wasp/packages/gpa"

// This event is introduced to avoid too-often consensus runs.
// They can produce more blocks than the PoV allows to confirm them.
// With this event the consensus will be proposed / started with
// a maximal rate defined by these events.
type inputCanPropose struct{}

func NewInputCanPropose() gpa.Input {
	return &inputCanPropose{}
}

func (inp *inputCanPropose) String() string {
	return "{cmtLog.inputCanPropose}"
}
