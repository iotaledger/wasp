package pipe

import (
	"testing"

	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

type dequeTestSM struct { // State machine for block WAL property based Rapid tests
	deque Deque[int]
	elems []int
}

var _ rapid.StateMachine = &dequeTestSM{}

func newDequeTestSM() *dequeTestSM {
	return &dequeTestSM{
		deque: NewDeque[int](),
		elems: make([]int, 0),
	}
}

func (dtsmT *dequeTestSM) Check(t *rapid.T) {
	dtsmT.invariantLength(t)
	dtsmT.invariantPeek(t)
	dtsmT.invariantAllElems(t)
}

func (dtsmT *dequeTestSM) invariantLength(t *rapid.T) {
	require.Equal(t, len(dtsmT.elems), dtsmT.deque.Length())
}

func (dtsmT *dequeTestSM) invariantPeek(t *rapid.T) {
	if len(dtsmT.elems) > 0 {
		require.Equal(t, dtsmT.elems[0], dtsmT.deque.PeekStart())
		require.Equal(t, dtsmT.elems[len(dtsmT.elems)-1], dtsmT.deque.PeekEnd())
		if len(dtsmT.elems) >= 3 {
			require.Equal(t, dtsmT.elems[:3], dtsmT.deque.PeekNStart(3))
			require.Equal(t, dtsmT.elems[len(dtsmT.elems)-3:], dtsmT.deque.PeekNEnd(3))
		} else {
			require.Equal(t, dtsmT.elems, dtsmT.deque.PeekNStart(3))
			require.Equal(t, dtsmT.elems, dtsmT.deque.PeekNEnd(3))
		}
	} else {
		require.Panics(t, func() { dtsmT.deque.PeekStart() })
		require.Panics(t, func() { dtsmT.deque.PeekEnd() })
		require.Equal(t, []int{}, dtsmT.deque.PeekNStart(3))
		require.Equal(t, []int{}, dtsmT.deque.PeekNEnd(3))
	}
	require.Equal(t, dtsmT.elems, dtsmT.deque.PeekNStart(len(dtsmT.elems)))
	require.Equal(t, dtsmT.elems, dtsmT.deque.PeekNEnd(len(dtsmT.elems)))
	require.Equal(t, dtsmT.elems, dtsmT.deque.PeekAll())
}

func (dtsmT *dequeTestSM) invariantAllElems(t *rapid.T) {
	for i := range dtsmT.elems {
		require.Equal(t, dtsmT.elems[i], dtsmT.deque.Get(i))
	}
}

func (dtsmT *dequeTestSM) AddStart(t *rapid.T) {
	elem := rapid.IntRange(0, 1000).Example()
	require.True(t, dtsmT.deque.AddStart(elem))
	dtsmT.elems = append([]int{elem}, dtsmT.elems...)
}

func (dtsmT *dequeTestSM) AddEnd(t *rapid.T) {
	elem := rapid.IntRange(0, 1000).Example()
	require.True(t, dtsmT.deque.AddEnd(elem))
	dtsmT.elems = append(dtsmT.elems, elem)
}

func (dtsmT *dequeTestSM) RemoveStart(t *rapid.T) {
	if len(dtsmT.elems) == 0 {
		t.Skip()
	}
	require.Equal(t, dtsmT.elems[0], dtsmT.deque.RemoveStart())
	dtsmT.elems = dtsmT.elems[1:]
}

func (dtsmT *dequeTestSM) RemoveEnd(t *rapid.T) {
	if len(dtsmT.elems) == 0 {
		t.Skip()
	}
	require.Equal(t, dtsmT.elems[len(dtsmT.elems)-1], dtsmT.deque.RemoveEnd())
	dtsmT.elems = dtsmT.elems[:len(dtsmT.elems)-1]
}

func (dtsmT *dequeTestSM) RemoveAt(t *rapid.T) {
	if len(dtsmT.elems) == 0 {
		t.Skip()
	}
	index := rapid.IntRange(-len(dtsmT.elems), len(dtsmT.elems)-1).Example()
	var posIndex int
	if index >= 0 {
		posIndex = index
	} else {
		posIndex = len(dtsmT.elems) + index
	}
	require.Equal(t, dtsmT.elems[posIndex], dtsmT.deque.RemoveAt(index))
	dtsmT.elems = append(dtsmT.elems[:posIndex], dtsmT.elems[posIndex+1:]...)
}

func TestDequePropBased(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		sm := newDequeTestSM()
		t.Repeat(rapid.StateMachineActions(sm))
	})
}
