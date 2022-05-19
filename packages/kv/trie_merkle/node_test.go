package trie_merkle

import (
	"testing"

	"github.com/iotaledger/wasp/packages/kv/trie"
	"github.com/stretchr/testify/require"
)

func TestNodeSerialization(t *testing.T) {
	t.Run("1", func(t *testing.T) {
		n := trie.NewNode(nil)
		n.ChildCommitments[1] = Model.NewVectorCommitment()
		n.ChildCommitments[6] = Model.NewVectorCommitment()
		n.ChildCommitments[255] = Model.NewVectorCommitment()

		bin := n.Bytes()
		nBack, err := trie.NodeFromBytes(Model, bin)
		require.NoError(t, err)
		require.EqualValues(t, n.Bytes(), nBack.Bytes())
	})
	t.Run("2", func(t *testing.T) {
		n := trie.NewNode(nil)
		n.Terminal = Model.NewTerminalCommitment()

		bin := n.Bytes()
		nBack, err := trie.NodeFromBytes(Model, bin)
		require.NoError(t, err)
		require.EqualValues(t, n.Bytes(), nBack.Bytes())
	})
}
