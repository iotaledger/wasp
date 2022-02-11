package trie

import (
	"bytes"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestNode(t *testing.T) {
	t.Run("base1", func(t *testing.T) {
		n := &Node{}
		var buf bytes.Buffer
		n.Write(&buf)
		t.Logf("size() = %d, size(serialize) = %d", Size(n), len(buf.Bytes()))
		require.EqualValues(t, Size(n), len(buf.Bytes()))

		nBack, err := MerkleTrieSetup.NodeFromBytes(buf.Bytes())
		require.NoError(t, err)
		require.EqualValues(t, buf.Bytes(), Bytes(nBack))

		h := MerkleTrieSetup.CommitToChildren(n)
		hBack := MerkleTrieSetup.CommitToChildren(nBack)
		require.EqualValues(t, h, hBack)
		t.Logf("commitment = %s", h)
	})
	t.Run("base short terminal", func(t *testing.T) {
		n := &Node{}
		n.pathFragment = []byte("kuku")
		n.terminalCommitment = MerkleTrieSetup.CommitToData([]byte("data"))

		var buf bytes.Buffer
		n.Write(&buf)
		t.Logf("size() = %d, size(serialize) = %d", Size(n), len(buf.Bytes()))
		require.EqualValues(t, Size(n), len(buf.Bytes()))

		nBack, err := MerkleTrieSetup.NodeFromBytes(buf.Bytes())
		require.NoError(t, err)
		require.EqualValues(t, buf.Bytes(), Bytes(nBack))

		h := MerkleTrieSetup.CommitToChildren(n)
		hBack := MerkleTrieSetup.CommitToChildren(nBack)
		require.EqualValues(t, h, hBack)
		t.Logf("commitment = %s", h)
	})
	t.Run("base long terminal", func(t *testing.T) {
		n := &Node{}
		n.pathFragment = []byte("kuku")
		n.terminalCommitment = MerkleTrieSetup.CommitToData([]byte(strings.Repeat("data", 1000)))

		var buf bytes.Buffer
		n.Write(&buf)
		t.Logf("size() = %d, size(serialize) = %d", Size(n), len(buf.Bytes()))
		require.EqualValues(t, Size(n), len(buf.Bytes()))

		nBack, err := MerkleTrieSetup.NodeFromBytes(buf.Bytes())
		require.NoError(t, err)
		require.EqualValues(t, buf.Bytes(), Bytes(nBack))

		h := MerkleTrieSetup.CommitToChildren(n)
		hBack := MerkleTrieSetup.CommitToChildren(nBack)
		require.EqualValues(t, h, hBack)
		t.Logf("commitment = %s", h)
	})
}

func TestTrie(t *testing.T) {
	k0 := []byte{}
	v0 := []byte("1234")
	k1 := []byte("a")
	v1 := []byte("5678")
	k2 := []byte("ab")
	v2 := []byte("ddd")
	t.Run("base", func(t *testing.T) {
		store := dict.New()
		tr := NewTrie(MerkleTrieSetup, store, nil)
		require.EqualValues(t, nil, tr.RootCommitment())

		tr.Update(k0, v0)
		t.Logf("root0 = %s", tr.RootCommitment())
		c := tr.RootCommitment()
		rootNode, ok := tr.GetNode(nil)
		require.True(t, ok)
		require.EqualValues(t, c, tr.setup.CommitToChildren(rootNode))

		tr.Update(k1, v1)
		t.Logf("root1 = %s", tr.RootCommitment())
		c = tr.RootCommitment()
		rootNode, ok = tr.GetNode(nil)
		require.True(t, ok)
		require.EqualValues(t, c, tr.setup.CommitToChildren(rootNode))

		tr.Update(k2, v2)
		t.Logf("root2 = %s", tr.RootCommitment())
		c = tr.RootCommitment()
		rootNode, ok = tr.GetNode(nil)
		require.True(t, ok)
		require.EqualValues(t, c, tr.setup.CommitToChildren(rootNode))

		tr.FlushCache(store)
		tr.ClearCache()
		require.EqualValues(t, 3, len(store))

		tr.Update(k2, v2)
		t.Logf("root = %s", tr.RootCommitment())
		rootNode, ok = tr.GetNode(nil)
		require.True(t, ok)
		require.EqualValues(t, c, tr.setup.CommitToChildren(rootNode))
		tr.Update(k2, v2)
	})
}
