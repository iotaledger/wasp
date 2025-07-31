package state

import (
	"testing"

	"pgregory.net/rapid"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/packages/trie"
)

func TestL1CommitmentConversionToBytes(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		trieRoot, err := trie.HashFromBytes(rapid.SliceOfN(rapid.Byte(), trie.HashSizeBytes, trie.HashSizeBytes).Draw(t, "trie root"))
		require.NoError(t, err)
		blockHash, err := NewBlockHash(rapid.SliceOfN(rapid.Byte(), BlockHashSize, BlockHashSize).Draw(t, "block hash"))
		require.NoError(t, err)
		l1Commitment := newL1Commitment(trieRoot, blockHash)

		l1CommitmentBack, err := NewL1CommitmentFromBytes(l1Commitment.Bytes())
		require.NoError(t, err)
		require.True(t, l1Commitment.TrieRoot().Equals(l1CommitmentBack.TrieRoot()))
		require.True(t, l1Commitment.BlockHash().Equals(l1CommitmentBack.BlockHash()))
		require.True(t, l1Commitment.Equals(l1CommitmentBack))
	})
}

/*func TestBlockHash(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		blockHashSlice := rapid.SliceOfN(rapid.Byte(), BlockHashSize, BlockHashSize).Draw(t, "block hash")
		var blockHash BlockHash
		copy(blockHash[:], blockHashSlice)
		blockHashString := blockHash.String()
		blockHashNew, err := BlockHashFromString(blockHashString)
		require.NoError(t, err)
		require.True(t, blockHash.Equals(blockHashNew))
	})
}*/
