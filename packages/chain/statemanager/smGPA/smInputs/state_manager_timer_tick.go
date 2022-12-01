package smInputs

import (
	"time"

	"github.com/iotaledger/wasp/packages/gpa"
)

type StateManagerTimerTick struct {
	time time.Time
}

var _ gpa.Input = &StateManagerTimerTick{}

func NewStateManagerTimerTick(timee time.Time) *StateManagerTimerTick {
	return &StateManagerTimerTick{time: timee}
}

func (smttT *StateManagerTimerTick) GetTime() time.Time {
	return smttT.time
}
