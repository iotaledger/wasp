package byteutils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConcatBytes(t *testing.T) {
	require.Equal(t, []byte{1, 2, 3, 4, 5, 7, 8}, ConcatBytes([]byte{1, 2, 3}, []byte{4, 5}, []byte{7, 8}))
	require.Equal(t, []byte{1, 2, 3, 4, 5, 7, 8}, ConcatBytes([]byte{1, 2, 3, 4, 5, 7, 8}))
}
