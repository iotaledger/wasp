package state

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestL1Commitment(t *testing.T) {
	sc := RandL1Commitment()

	data := sc.Bytes()
	scBack, err := L1CommitmentFromBytes(data)
	require.NoError(t, err)
	require.True(t, sc.TrieRoot.Equals(scBack.TrieRoot))
	require.EqualValues(t, sc.BlockHash, scBack.BlockHash)
}
