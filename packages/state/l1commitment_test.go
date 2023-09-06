package state

import (
	"testing"

	"pgregory.net/rapid"

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

func TestBlockHash(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		blockHashSlice := rapid.SliceOfN(rapid.Byte(), BlockHashSize, BlockHashSize).Draw(t, "block hash")
		var blockHash BlockHash
		copy(blockHash[:], blockHashSlice)
		blockHashString := blockHash.String()
		blockHashNew, err := BlockHashFromString(blockHashString)
		require.NoError(t, err)
		require.True(t, blockHash.Equals(blockHashNew))
	})
}
