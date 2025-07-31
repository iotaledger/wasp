package pipe_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/packages/util/pipe"
)

func TestSimple(t *testing.T) {
	deque := pipe.NewDeque[int]()
	require.Equal(t, 0, deque.Length())
	for i := 0; i < 100; i++ {
		elem := i * 2
		if i%2 == 0 {
			require.True(t, deque.AddStart(elem))
			require.Equal(t, elem, deque.PeekStart())
		} else {
			require.True(t, deque.AddEnd(elem))
			require.Equal(t, elem, deque.PeekEnd())
		}
		require.Equal(t, i+1, deque.Length())
	}
	for i := 99; i >= 0; i-- {
		elem := i * 2
		if i%2 == 0 {
			require.Equal(t, elem, deque.RemoveStart())
		} else {
			require.Equal(t, elem, deque.RemoveEnd())
		}
		require.Equal(t, i, deque.Length())
	}
}
