package trie

import (
	"github.com/iotaledger/wasp/packages/kv"
)

type TrieSetup struct {
	NewVectorCommitment   func() VectorCommitment
	NewTerminalCommitment func() TerminalCommitment
	CommitToChildren      func(*Node) VectorCommitment
	CommitToData          func([]byte) TerminalCommitment
	UpdateKey             func(t *trie, path []byte, pathPosition int, updateCommitment *VectorCommitment, terminal TerminalCommitment)
}

type trie struct {
	setup          *TrieSetup
	store          kv.KVMustReader
	rootCommitment VectorCommitment
	nodeCache      map[kv.Key]*Node
}

func NewTrie(setup *TrieSetup, store kv.KVMustReader, rootCommitment VectorCommitment) *trie {
	return &trie{
		setup:          setup,
		store:          store,
		rootCommitment: rootCommitment,
		nodeCache:      make(map[kv.Key]*Node),
	}
}

func (t *trie) RootCommitment() VectorCommitment {
	return t.rootCommitment
}

func (t *trie) GetNode(key []byte) (*Node, bool) {
	k := kv.Key(key)
	node, ok := t.nodeCache[k]
	if ok {
		return node, true
	}
	nodeBin := t.store.MustGet(k)
	if nodeBin == nil {
		return nil, false
	}
	node, err := t.setup.NodeFromBytes(nodeBin)
	assert(err == nil, err)

	t.nodeCache[k] = node
	return node, true

}

// assumes that node with the key does not exist
func (t *trie) mustNewNode(key []byte) *Node {
	t.nodeCache[kv.Key(key)] = &Node{}
	return t.nodeCache[kv.Key(key)]
}

func (t *trie) Update(key, value []byte) {
	if len(value) == 0 {
		// empty value means no value and no entry
		return
	}
	terminal := t.setup.CommitToData(value)
	t.setup.UpdateKey(t, key, 0, &t.rootCommitment, terminal)
}

func (t *trie) FlushCache(store kv.KVStore) {
	for k, v := range t.nodeCache {
		store.Set(k, Bytes(v))
	}
}

func (t *trie) ClearCache() {
	t.nodeCache = make(map[kv.Key]*Node)
}
