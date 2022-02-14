package trie

import (
	"bytes"
	"github.com/iotaledger/wasp/packages/kv"
)

// CommitmentLogic abstracts 256+ Trie logic from the commitment logic/cryptography
type CommitmentLogic interface {
	NewVectorCommitment() VectorCommitment
	NewTerminalCommitment() TerminalCommitment
	CommitToNode(*Node) VectorCommitment
	CommitToData([]byte) TerminalCommitment
	UpdateNodeCommitment(*Node) VectorCommitment // returns delta. Return nil if deletion
	UpdateCommitment(update *VectorCommitment, delta VectorCommitment)
}

type Trie struct {
	setup          CommitmentLogic
	store          kv.KVMustReader
	rootCommitment VectorCommitment
	nodeCache      map[kv.Key]*Node
}

type ProofPath struct {
	Key  []byte
	Path []*ProofPathElement
}

type ProofPathElement struct {
	Key        []byte
	Node       *Node
	ChildIndex int
}

func NewTrie(setup CommitmentLogic, store kv.KVMustReader, rootCommitment VectorCommitment) *Trie {
	return &Trie{
		setup:          setup,
		store:          store,
		rootCommitment: rootCommitment,
		nodeCache:      make(map[kv.Key]*Node),
	}
}

func (t *Trie) RootCommitment() VectorCommitment {
	return t.rootCommitment
}

// getNode takes node from the cache of fetches it from kv store
func (t *Trie) GetNode(key []byte) (*Node, bool) {
	k := kv.Key(key)
	node, ok := t.nodeCache[k]
	if ok {
		return node, true
	}
	nodeBin := t.store.MustGet(k)
	if nodeBin == nil {
		return nil, false
	}
	node, err := NodeFromBytes(t.setup, nodeBin)
	assert(err == nil, err)

	t.nodeCache[k] = node
	return node, true

}

func (t *Trie) CommitToNode(n *Node) VectorCommitment {
	return t.setup.CommitToNode(n)
}

func (t *Trie) FlushCache(store kv.KVStore) {
	for k, v := range t.nodeCache {
		if !v.IsEmpty() {
			store.Set(k, Bytes(v))
		} else {
			store.Del(k)
		}
	}
}

func (t *Trie) ClearCache() {
	t.nodeCache = make(map[kv.Key]*Node)
}

// newTerminalNode assumes Key does not exist in the Trie
func (t *Trie) newTerminalNode(key, pathFragment []byte, newTerminal TerminalCommitment) *Node {
	ret := NewNode(pathFragment)
	ret.NewTerminal = newTerminal
	t.nodeCache[kv.Key(key)] = ret
	return ret
}

func (t *Trie) newNodeCopy(key, pathFragment []byte, copyFrom *Node) *Node {
	ret := *copyFrom
	ret.PathFragment = pathFragment
	t.nodeCache[kv.Key(key)] = &ret
	return &ret
}

// Update updates Trie with the Key/value
func (t *Trie) Update(key []byte, value []byte) {
	c := t.setup.CommitToData(value)
	t.updateKey(key, 0, c)
}

// Delete deletes Key/value from the Trie
func (t *Trie) Delete(key []byte) {
	t.Update(key, nil)
}

// updateKey updates tree (recursively) by adding or updating terminal commitment at specified Key
// - 'path' the Key of the terminal value
// - 'pathPosition' is the position in the path the current node's Key starts: Key = path[:pathPosition]
// - 'terminal' is the new commitment to the value under Key 'path'. nil means terminal deletion
// - returns the node under the Key = path[:pathPosition] or nil in case of terminal deletion
func (t *Trie) updateKey(path []byte, pathPosition int, terminal TerminalCommitment) *Node {
	assert(pathPosition <= len(path), "pathPosition <= len(path)")
	if len(path) == 0 {
		path = []byte{}
	}
	key := path[:pathPosition]
	tail := path[pathPosition:]
	// looking up for the node with the Key (with caching)
	node, ok := t.GetNode(key)
	if !ok {
		if terminal == nil {
			// in case of deletion do nothing
			return nil
		}
		// node for the path[:pathPosition] does not exist. Create a new one with the terminal value only
		return t.newTerminalNode(key, tail, terminal)
	}
	// node for the Key exists. Find common prefix between tail of the path and path fragment
	prefix := commonPrefix(node.PathFragment, tail)
	assert(len(prefix) <= len(node.PathFragment), "len(prefix)<= len(node.pathFragment)")
	// the following parameters define how it goes:
	// - len(path)
	// - pathPosition
	// - len(node.pathFragment)
	// - len(prefix)
	nextPathPosition := pathPosition + len(prefix)
	assert(nextPathPosition <= len(path), "nextPathPosition <= len(path)")

	if len(prefix) == len(node.PathFragment) {
		// pathFragment is part of the path. No need for a fork, continue the path
		if nextPathPosition == len(path) {
			// reached the terminal value on this node. In case of deletion the newTerminal will become nil
			node.NewTerminal = terminal
			return node
		}
		assert(nextPathPosition < len(path), "nextPathPosition < len(path)")
		// didn't reach the end of the path
		// choose the direction and continue down the path of the child
		childIndex := path[nextPathPosition]
		// recursively update the rest of the path
		node.ModifiedChildren[childIndex] = t.updateKey(path, nextPathPosition+1, terminal)
		return node
	}
	assert(len(prefix) < len(node.PathFragment), "len(prefix) < len(node.pathFragment)")

	if terminal == nil {
		// node not found, do nothing in case of deletion
		return nil
	}
	// split the pathFragment. The continued branch is part of the fragment
	// Key of the next node starts at the next position after current Key plus prefix
	keyContinue := make([]byte, pathPosition+len(prefix)+1)
	copy(keyContinue, path)
	keyContinue[len(keyContinue)-1] = node.PathFragment[len(prefix)]
	// add child index to the end of the keyContinue
	childIndexContinue := keyContinue[len(keyContinue)-1]
	// create new node on keyContinue, move everything from old to the new node and adjust the path fragment
	insertNode := t.newNodeCopy(keyContinue, node.PathFragment[len(prefix)+1:], node)
	// clear the old one and adjust path fragment. Continue with 1 child, the new node
	node.Children = make(map[uint8]VectorCommitment)
	node.ModifiedChildren = make(map[uint8]*Node)
	node.PathFragment = prefix
	node.ModifiedChildren[childIndexContinue] = insertNode
	node.Terminal = nil
	node.NewTerminal = nil

	if pathPosition+len(prefix) == len(path) {
		// no need for the new node
		node.NewTerminal = terminal
	} else {
		// create the new node
		keyFork := path[:pathPosition+len(prefix)+1]
		childForkIndex := keyFork[len(keyFork)-1]
		assert(len(keyContinue) == len(keyFork), "len(keyContinue)==len(keyFork)")
		node.ModifiedChildren[childForkIndex] = t.newTerminalNode(keyFork, path[len(keyFork):], terminal)
	}
	return node
}

// Commit calculates a new root commitment value from the cache and commits all mutations in the cached nodes
// Doesn't delete cached nodes
func (t *Trie) Commit() {
	root, ok := t.GetNode(nil)
	if !ok {
		t.rootCommitment = nil
		return
	}
	deltaC := t.setup.UpdateNodeCommitment(root)
	t.setup.UpdateCommitment(&t.rootCommitment, deltaC)
}

// ProofPath returns generic proof path
func (t *Trie) ProofPath(key []byte) *ProofPath {
	if len(key) == 0 {
		key = []byte{}
	}
	ret := &ProofPath{
		Key:  key,
		Path: make([]*ProofPathElement, 0),
	}
	t.path(0, ret)
	return ret
}

func (t *Trie) path(pathPosition int, ret *ProofPath) {
	assert(pathPosition <= len(ret.Key), "pathPosition <= len(path)")
	key := ret.Key[:pathPosition]
	// looking up for the node with the Key (with caching)
	node, ok := t.GetNode(key)
	if !ok {
		// Key is absent. Should not happen, except when empty trie
		return
	}
	elem := &ProofPathElement{
		Key:  key,
		Node: node,
	}
	tail := ret.Key[pathPosition:]
	if !bytes.HasPrefix(tail, node.PathFragment) {
		elem.ChildIndex = 257
		return
	}
	if bytes.Equal(tail, node.PathFragment) {
		elem.ChildIndex = 256
		return
	}
	indexPos := pathPosition + len(node.PathFragment)
	assert(indexPos < len(ret.Key), "assertion: pathPosition+len(node.pathFragment)<=len(ret.Key)")
	elem.ChildIndex = int(ret.Key[indexPos])
	assert(node.Children[byte(elem.ChildIndex)] != nil, "assertion: node.Children[indexPos] != nil")
	t.path(indexPos+1, ret)
}

func commonPrefix(b1, b2 []byte) []byte {
	ret := make([]byte, 0)
	for i := 0; i < len(b1) && i < len(b2); i++ {
		if b1[i] != b2[i] {
			break
		}
		ret = append(ret, b1[i])
	}
	return ret
}

func assert(cond bool, err interface{}) {
	if !cond {
		panic(err)
	}
}
