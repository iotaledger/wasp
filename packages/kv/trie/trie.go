package trie

import (
	"bytes"
	"github.com/iotaledger/wasp/packages/kv"
)

// CommitmentModel abstracts 256+ Trie logic from the commitment logic/cryptography
type CommitmentModel interface {
	NewVectorCommitment() VCommitment
	NewTerminalCommitment() TCommitment
	CommitToNode(*Node) VCommitment
	CommitToData([]byte) TCommitment
	UpdateNodeCommitment(*Node) VCommitment // returns delta. Return nil if deletion
}

type Trie struct {
	model     CommitmentModel
	store     kv.KVMustReader
	nodeCache map[kv.Key]*Node
}

type ProofGeneric struct {
	Key  []byte
	Path []*ProofGenericElement
}

type ProofGenericElement struct {
	Key        []byte
	Node       *Node
	ChildIndex int
}

func New(model CommitmentModel, store kv.KVMustReader) *Trie {
	ret := &Trie{
		model:     model,
		store:     store,
		nodeCache: make(map[kv.Key]*Node),
	}
	return ret
}

func (t *Trie) Clone() *Trie {
	ret := &Trie{
		model:     t.model,
		store:     t.store,
		nodeCache: make(map[kv.Key]*Node),
	}
	for k, v := range t.nodeCache {
		ret.nodeCache[k] = v.Clone()
	}
	return ret
}

func (t *Trie) RootCommitment() VCommitment {
	n, ok := t.GetNode(nil)
	if !ok {
		return nil
	}
	return t.CommitToNode(n)
}

// GetNode takes node from the cache of fetches it from kv store
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
	node, err := NodeFromBytes(t.model, nodeBin)
	assert(err == nil, err)

	t.nodeCache[k] = node
	return node, true

}

func (t *Trie) CommitToNode(n *Node) VCommitment {
	return t.model.CommitToNode(n)
}

func (t *Trie) ApplyMutations(store kv.KVWriter) {
	for k, v := range t.nodeCache {
		if !v.IsEmpty() {
			store.Set(k, MustBytes(v))
		} else {
			store.Del(k)
		}
	}
}

func (t *Trie) ClearCache() {
	t.nodeCache = make(map[kv.Key]*Node)
}

// newTerminalNode assumes Key does not exist in the Trie
func (t *Trie) newTerminalNode(key, pathFragment []byte, newTerminal TCommitment) *Node {
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
	c := t.model.CommitToData(value)
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
func (t *Trie) updateKey(path []byte, pathPosition int, terminal TCommitment) *Node {
	assert(pathPosition <= len(path), "pathPosition <= len(path)")
	if len(path) == 0 {
		path = []byte{}
	}
	key := path[:pathPosition]
	tail := path[pathPosition:]
	// looking up for the node with the key (with caching)
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
	node.Children = make(map[uint8]VCommitment)
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
		return
	}
	t.model.UpdateNodeCommitment(root)
}

// ProofGeneric returns generic proof path. Contains references trie node cache.
// Should be immediately converted into the specific proof model independent of the trie
// Normally only called by the model
func (t *Trie) ProofGeneric(key []byte) *ProofGeneric {
	if len(key) == 0 {
		key = []byte{}
	}
	ret := &ProofGeneric{
		Key:  key,
		Path: make([]*ProofGenericElement, 0),
	}
	t.path(0, ret)
	return ret
}

func (t *Trie) path(pathPosition int, ret *ProofGeneric) {
	assert(pathPosition <= len(ret.Key), "pathPosition <= len(path)")
	key := ret.Key[:pathPosition]
	// looking up for the node with the Key (with caching)
	node, ok := t.GetNode(key)
	if !ok {
		// Key is absent. Should not happen, except when empty trie
		return
	}
	elem := &ProofGenericElement{
		Key:  key,
		Node: node,
	}
	ret.Path = append(ret.Path, elem)
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
	if node.Children[byte(elem.ChildIndex)] == nil {
		// no way further
		return
	}
	t.path(indexPos+1, ret)
}

func (t *Trie) VectorCommitmentFromBytes(data []byte) (VCommitment, error) {
	ret := t.model.NewVectorCommitment()
	if err := ret.Read(bytes.NewReader(data)); err != nil {
		return nil, err
	}
	return ret, nil
}
