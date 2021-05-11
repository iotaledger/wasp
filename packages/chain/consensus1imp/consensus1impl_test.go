package consensus1imp

import (
	"testing"
)

func TestConsensusEnv1(t *testing.T) {
	t.Run("wait index", func(t *testing.T) {
		env, _ := NewMockedEnv(t, 4, 3, true)
		env.StartTimers()
		env.eventStateTransition()
		env.WaitStateIndex(0)
	})
	t.Run("wait timer tick", func(t *testing.T) {
		env, _ := NewMockedEnv(t, 4, 3, true)
		env.StartTimers()
		env.eventStateTransition()
		env.WaitTimerTick(43)
	})
}
