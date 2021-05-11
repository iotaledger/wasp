package consensus1imp

import (
	"testing"
)

func TestConsensusEnv(t *testing.T) {
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

func TestConsensusPostRequest(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	t.Run("post 1", func(t *testing.T) {
		env, _ := NewMockedEnv(t, 4, 3, true)
		env.StartTimers()
		env.eventStateTransition()
		env.postDummyRequest()
		env.WaitTimerTick(10)
	})

	t.Run("post 1 randomize", func(t *testing.T) {
		env, _ := NewMockedEnv(t, 4, 3, true)
		env.StartTimers()
		env.eventStateTransition()
		env.postDummyRequest(true)
		env.WaitTimerTick(100)
	})

}
