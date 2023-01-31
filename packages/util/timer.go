package util

import (
	"fmt"
	"time"
)

type timerStep struct {
	name     string
	start    time.Time
	duration time.Duration
}
type Timer struct {
	steps []*timerStep
}

func NewTimer() *Timer {
	return &Timer{
		steps: []*timerStep{
			{name: "pending", start: time.Now()},
		},
	}
}

func (t *Timer) Duration() time.Duration {
	return time.Since(t.steps[0].start)
}

func (t *Timer) Step(name string) {
	t.Done(name)
	t.steps = append(t.steps, t.newStep())
}

func (t *Timer) lastStep() *timerStep {
	return t.steps[len(t.steps)-1]
}

func (t *Timer) newStep() *timerStep {
	return &timerStep{name: "pending", start: time.Now()}
}

func (t *Timer) Done(name string) {
	lastStep := t.lastStep()
	if lastStep.duration == 0 {
		lastStep.name = name
		lastStep.duration = time.Since(lastStep.start)
	}
}

func (t *Timer) String() string {
	t.Done("last")
	if len(t.steps) == 1 {
		return fmt.Sprintf("Total: %v", t.Duration())
	}
	stepsStr := ""
	for _, st := range t.steps {
		stepsStr += fmt.Sprintf(", %v=%v", st.name, st.duration)
	}
	return fmt.Sprintf("Total: %v%s", t.Duration(), stepsStr)
}
