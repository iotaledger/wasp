package pipe

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSimple(t *testing.T) {
	deque := NewDeque[int]()
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
