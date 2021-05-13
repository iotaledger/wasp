package consensus1imp

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConsensusEnv(t *testing.T) {
	t.Run("wait index", func(t *testing.T) {
		env, _ := NewMockedEnv(t, 4, 3, true)
		env.StartTimers()
		env.eventStateTransition()
		err := env.WaitStateIndex(4, 0)
		require.NoError(t, err)
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
		err := env.WaitStateIndex(3, 1)
		require.NoError(t, err)
	})
	t.Run("post 1 randomize", func(t *testing.T) {
		env, _ := NewMockedEnv(t, 4, 3, true)
		env.StartTimers()
		env.eventStateTransition()
		env.postDummyRequest(true)
		err := env.WaitStateIndex(3, 1)
		require.NoError(t, err)
	})

}
