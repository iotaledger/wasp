package state

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestL1Commitment(t *testing.T) {
	sc := PseudoRandL1Commitment()

	data := sc.Bytes()
	scBack, err := L1CommitmentFromBytes(data)
	require.NoError(t, err)
	require.Equal(t, sc.TrieRoot(), scBack.TrieRoot())
	require.Equal(t, sc.BlockHash(), scBack.BlockHash())
}
