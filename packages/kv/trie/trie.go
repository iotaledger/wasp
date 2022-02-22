package trie

import (
	"bytes"
	"fmt"
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
	Key    []byte
	Path   []ProofGenericElement
	Ending ProofEndingCode
}

type ProofGenericElement struct {
	Key  []byte
	Node *Node
}

type ProofEndingCode byte

const (
	EndingTerminal = iota
	EndingSplit
	EndingExtend
)

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
	ret.ModifiedTerminal = newTerminal
	t.nodeCache[kv.Key(key)] = ret
	return ret
}

func (t *Trie) newNodeCopy(key, pathFragment []byte, copyFrom *Node) *Node {
	ret := *copyFrom
	ret.PathFragment = pathFragment
	t.nodeCache[kv.Key(key)] = &ret
	return &ret
}

// Delete deletes Key/value from the Trie
func (t *Trie) Delete(key []byte) {
	t.Update(key, nil)
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
	p, _, _, ending := t.path(key, 0)
	return &ProofGeneric{
		Key:    key,
		Path:   p,
		Ending: ending,
	}
}

// Update updates Trie with the key/value.
// value == nil means deletion
func (t *Trie) Update(key []byte, value []byte) {
	c := t.model.CommitToData(value)
	proof, lastKey, lastCommonPrefix, ending := t.path(key, 0)
	if len(proof) == 0 {
		if c != nil {
			t.newTerminalNode(nil, key, c)
		}
		return

	}
	last := proof[len(proof)-1].Node
	switch ending {
	case EndingTerminal:
		last.ModifiedTerminal = c

	case EndingExtend:
		if c == nil {
			break
		}
		childIndexPosition := len(lastKey) + len(lastCommonPrefix)
		assert(childIndexPosition < len(key), "childPosition < len(key)")
		childIndex := key[childIndexPosition]
		assert(last.Children[childIndex] == nil, "last.Children[key[childPosition]] == nil")
		assert(last.ModifiedChildren[childIndex] == nil, "last.ModifiedChildren[key[childPosition]] == nil")
		last.ModifiedChildren[childIndex] = t.newTerminalNode(key[:childIndexPosition+1], key[childIndexPosition+1:], c)

	case EndingSplit:
		if c == nil {
			break
		}
		childPosition := len(lastKey) + len(lastCommonPrefix)
		assert(childPosition <= len(key), "childPosition < len(key)")
		keyContinue := make([]byte, childPosition+1)
		copy(keyContinue, key)
		splitChildIndex := len(lastCommonPrefix)
		assert(splitChildIndex < len(last.PathFragment), "splitChildIndex<len(last.Node.PathFragment)")
		childContinue := last.PathFragment[splitChildIndex]
		keyContinue[len(keyContinue)-1] = childContinue

		// create new node on keyContinue, move everything from old to the new node and adjust the path fragment
		insertNode := t.newNodeCopy(keyContinue, last.PathFragment[splitChildIndex+1:], last)
		// clear the old one and adjust path fragment. Continue with 1 child, the new node
		last.Children = make(map[uint8]VCommitment)
		last.ModifiedChildren = make(map[uint8]*Node)
		last.PathFragment = lastCommonPrefix
		last.ModifiedChildren[childContinue] = insertNode
		last.Terminal = nil
		last.ModifiedTerminal = nil
		// insert terminal
		if childPosition == len(key) {
			// no need for the new node
			last.ModifiedTerminal = c
		} else {
			// create a new node
			keyFork := key[:len(keyContinue)]
			childForkIndex := keyFork[len(keyFork)-1]
			assert(int(childForkIndex) != splitChildIndex, "childForkIndex != splitChildIndex")
			last.ModifiedChildren[childForkIndex] = t.newTerminalNode(keyFork, key[len(keyFork):], c)
		}

	default:
		panic("inconsistency: unknown path ending code")
	}
	// update commitment path back to root
	for i := len(proof) - 2; i >= 0; i-- {
		k := proof[i+1].Key
		childIndex := k[len(k)-1]
		proof[i].Node.ModifiedChildren[childIndex] = proof[i+1].Node
	}
}

// returns key of the last node and common prefix with the fragment
func (t *Trie) path(path []byte, pathPosition int) ([]ProofGenericElement, []byte, []byte, ProofEndingCode) {
	node, ok := t.GetNode(nil)
	if !ok {
		return nil, nil, nil, 0
	}

	proof := []ProofGenericElement{{Key: nil, Node: node}}
	key := path[:pathPosition]
	tail := path[pathPosition:]

	for {
		assert(len(key) <= len(path), "pathPosition<=len(path)")
		if bytes.Equal(tail, node.PathFragment) {
			return proof, nil, nil, EndingTerminal
		}
		prefix := commonPrefix(tail, node.PathFragment)

		if len(prefix) < len(node.PathFragment) {
			return proof, key, prefix, EndingSplit
		}
		// continue with the path. 2 options:
		// - it ends here
		// - it goes further
		assert(len(prefix) == len(node.PathFragment), "len(prefix)==len(node.PathFragment)")
		childIndexPosition := len(key) + len(prefix)
		assert(childIndexPosition < len(path), "childIndexPosition<len(path)")
		childIndex := path[childIndexPosition]
		if node.Children[childIndex] == nil && node.ModifiedChildren[childIndex] == nil {
			// if there are no commitment to the child at the position, it means trie must be extended at this point
			return proof, key, prefix, EndingExtend
		}
		// it means we continue the branch of commitment
		key = path[:childIndexPosition+1]
		tail = path[childIndexPosition+1:]
		node, ok = t.GetNode(key)
		if !ok {
			panic(fmt.Sprintf("inconsistency: trie key not found: '%s'", string(key)))
		}
		proof = append(proof, ProofGenericElement{Key: key, Node: node})
	}
}

func (t *Trie) VectorCommitmentFromBytes(data []byte) (VCommitment, error) {
	ret := t.model.NewVectorCommitment()
	if err := ret.Read(bytes.NewReader(data)); err != nil {
		return nil, err
	}
	return ret, nil
}

func EqualCommitments(c1, c2 CommitmentBase) bool {
	if c1 == c2 {
		return true
	}
	if c1 == nil || c2 == nil {
		return false
	}
	return c1.Equal(c2)
}
