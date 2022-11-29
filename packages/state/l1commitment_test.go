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
	require.Equal(t, sc.GetTrieRoot(), scBack.GetTrieRoot())
	require.Equal(t, sc.GetBlockHash(), scBack.GetBlockHash())
}
