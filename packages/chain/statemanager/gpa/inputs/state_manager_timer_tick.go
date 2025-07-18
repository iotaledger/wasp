package inputs

import (
	"time"

	"github.com/iotaledger/wasp/v2/packages/gpa"
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
