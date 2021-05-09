package consensus1imp

import (
	"testing"
	"time"
)

func TestConsensus1ImplEnv(t *testing.T) {
	env, _ := NewMockedEnv(t, 4, 3, true)
	env.eventStateTransition()
	time.Sleep(100 * time.Millisecond)
}
