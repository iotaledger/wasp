package trie

import (
	"github.com/iotaledger/wasp/packages/kv"
)

type TrieSetup struct {
	NewVectorCommitment   func() VectorCommitment
	NewTerminalCommitment func() TerminalCommitment
	CommitToChildren      func(*Node) VectorCommitment
	CommitToData          func([]byte) TerminalCommitment
	UpdateNodeCommitment  func(*Node) VectorCommitment // returns delta
	UpdateCommitment      func(update *VectorCommitment, delta VectorCommitment)
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

// GetNode takes node from the cache of fetches it from kv store
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

func (t *trie) FlushCache(store kv.KVStore) {
	for k, v := range t.nodeCache {
		store.Set(k, Bytes(v))
	}
}

func (t *trie) ClearCache() {
	t.nodeCache = make(map[kv.Key]*Node)
}

// newTerminalNode assumes key does not exist in the trie
func (t *trie) newTerminalNode(key, pathFragment []byte, newTerminal TerminalCommitment) *Node {
	ret := NewNode()
	ret.pathFragment = pathFragment
	ret.newTerminal = newTerminal
	t.nodeCache[kv.Key(key)] = ret
	return ret
}

func (t *trie) Update(key []byte, value []byte) {
	c := t.setup.CommitToData(value)
	t.updateKey(key, 0, c)
}

// updateKey updates tree (recursively) by adding or updating terminal commitment at specified key
// - 'path' the key of the terminal value
// - 'pathPosition' is the position in the path the current node's key starts: key = path[:pathPosition]
// - 'terminal' is the new commitment to the value under key 'path'
func (t *trie) updateKey(path []byte, pathPosition int, terminal TerminalCommitment) *Node {
	assert(pathPosition <= len(path), "pathPosition <= len(path)")
	if len(path) == 0 {
		path = []byte{}
	}
	key := path[:pathPosition]
	tail := path[pathPosition:]
	// looking up for the node with the key (with caching)
	node, ok := t.GetNode(key)
	if !ok {
		// node for the path[:pathPosition] does not exist. Create a new one with the terminal value only
		return t.newTerminalNode(key, tail, terminal)
	}
	// node for the key exists
	// find common prefix between tail of the
	prefix := commonPrefix(node.pathFragment, tail)
	assert(len(prefix) <= len(node.pathFragment), "len(prefix)<= len(node.pathFragment)")
	// the following parameters define how it goes:
	// - len(path)
	// - pathPosition
	// - len(node.pathFragment)
	// - len(prefix)
	nextPathPosition := pathPosition + len(prefix)
	assert(nextPathPosition <= len(path), "nextPathPosition <= len(path)")

	if len(prefix) == len(node.pathFragment) {
		// pathFragment is part of the path. No need for a fork, continue the path
		if nextPathPosition == len(path) {
			// reached the terminal value on this node
			node.newTerminal = terminal
			return node
		}
		assert(nextPathPosition < len(path), "nextPathPosition < len(path)")
		// didn't reach the end of the path
		// choose the direction and continue down the path of the child
		childIndex := path[nextPathPosition]
		// recursively update the rest of the path
		node.modifiedChildren[childIndex] = t.updateKey(path, nextPathPosition+1, terminal)
		return node
	}
	assert(len(prefix) < len(node.pathFragment), "len(prefix) < len(node.pathFragment)")

	// split the pathFragment. The continued branch is part of the fragment
	// key of the next node starts at the next position after current key plus prefix
	keyContinue := make([]byte, pathPosition+len(prefix)+1)
	copy(keyContinue, path)
	keyContinue[len(keyContinue)-1] = node.pathFragment[len(prefix)]
	// add child index to the end of the keyContinue
	childIndexContinue := keyContinue[len(keyContinue)-1]
	// create new node on keyContinue, move everything from old to the new node and adjust the path fragment
	newNode := NewNode()
	newNode.pathFragment = node.pathFragment[len(prefix)+1:]
	newNode.children = node.children
	newNode.modifiedChildren = node.modifiedChildren
	newNode.terminalCommitment = node.terminalCommitment
	newNode.newTerminal = node.newTerminal
	t.nodeCache[kv.Key(keyContinue)] = newNode
	// clear the old one and adjust path fragment. Continue with 1 child, the new node
	node.pathFragment = prefix
	node.children = make(map[uint8]VectorCommitment)
	node.modifiedChildren = make(map[uint8]*Node)
	node.modifiedChildren[childIndexContinue] = newNode
	node.terminalCommitment = nil
	node.newTerminal = nil

	if pathPosition+len(prefix) == len(path) {
		// no need for the new node
		node.newTerminal = terminal
	} else {
		// create the new node
		keyFork := path[:pathPosition+len(prefix)+1]
		assert(len(keyContinue) == len(keyFork), "len(keyContinue)==len(keyFork)")
		nodeFork := NewNode()
		nodeFork.pathFragment = path[len(keyFork):]
		nodeFork.newTerminal = terminal
		childForkIndex := keyFork[len(keyFork)-1]
		node.modifiedChildren[childForkIndex] = nodeFork
		t.nodeCache[kv.Key(keyFork)] = nodeFork
	}
	return node
}

// Commit calculates a new root commitment value from the cache and commits all mutations.
// Flush() writes the cache to kv store
func (t *trie) Commit() {
	root, ok := t.GetNode(nil)
	if !ok {
		t.rootCommitment = nil
		return
	}
	deltaC := t.setup.UpdateNodeCommitment(root)
	t.setup.UpdateCommitment(&t.rootCommitment, deltaC)
}
