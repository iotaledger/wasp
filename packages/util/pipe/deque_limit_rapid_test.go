package pipe

import (
	"testing"

	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

type limitedDequeTestSM struct { // State machine for block WAL property based Rapid tests
	*dequeTestSM
	limit int
}

var _ rapid.StateMachine = &limitedDequeTestSM{}

func newLimitedDequeTestSM(limit int) *limitedDequeTestSM {
	return &limitedDequeTestSM{
		dequeTestSM: &dequeTestSM{
			deque: NewLimitedDeque[int](limit),
			elems: make([]int, 0),
		},
		limit: limit,
	}
}

func (dtsmT *limitedDequeTestSM) AddStart(t *rapid.T) {
	if len(dtsmT.elems) >= dtsmT.limit {
		require.False(t, dtsmT.deque.AddStart(rapid.IntRange(0, 1000).Example()))
	} else {
		dtsmT.dequeTestSM.AddStart(t)
	}
}

func (dtsmT *limitedDequeTestSM) AddEnd(t *rapid.T) {
	if len(dtsmT.elems) >= dtsmT.limit {
		require.False(t, dtsmT.deque.AddEnd(rapid.IntRange(0, 1000).Example()))
	} else {
		dtsmT.dequeTestSM.AddEnd(t)
	}
}

func TestLimitedDequePropBased(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		sm := newLimitedDequeTestSM(10)
		t.Repeat(rapid.StateMachineActions(sm))
	})
}
