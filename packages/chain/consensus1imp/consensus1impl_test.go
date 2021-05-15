package consensus1imp

import (
	"testing"
	"time"

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
		env.postDummyRequests(1)
		err := env.WaitStateIndex(3, 1, 5*time.Second)
		require.NoError(t, err)
	})
	t.Run("post 1 randomize", func(t *testing.T) {
		env, _ := NewMockedEnv(t, 4, 3, true)
		env.StartTimers()
		env.eventStateTransition()
		env.postDummyRequests(1, true)
		err := env.WaitStateIndex(3, 1)
		require.NoError(t, err)
	})
	t.Run("post 10 requests", func(t *testing.T) {
		env, _ := NewMockedEnv(t, 4, 3, true)
		env.StartTimers()
		env.eventStateTransition()
		env.postDummyRequests(10)
		err := env.WaitMempool(10, 3, 5*time.Second)
		require.NoError(t, err)
	})
	t.Run("post 10 requests randomized", func(t *testing.T) {
		env, _ := NewMockedEnv(t, 4, 3, false)
		env.StartTimers()
		env.eventStateTransition()
		env.postDummyRequests(10, true)
		err := env.WaitMempool(10, 3, 5*time.Second)
		require.NoError(t, err)
	})

}
