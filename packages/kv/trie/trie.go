package trie

import (
	"bytes"
	"fmt"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

// CommitmentModel abstracts 256+ Trie logic from the commitment logic/cryptography
type CommitmentModel interface {
	NewVectorCommitment() VCommitment
	NewTerminalCommitment() TCommitment
	CommitToNode(*Node) VCommitment
	CommitToData([]byte) TCommitment
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

func (e ProofEndingCode) String() string {
	switch e {
	case EndingTerminal:
		return "EndingTerminal"
	case EndingSplit:
		return "EndingSplit"
	case EndingExtend:
		return "EndingExtend"
	default:
		panic("wrong ending code")
	}
}

func New(model CommitmentModel, store kv.KVMustReader) *Trie {
	ret := &Trie{
		model:     model,
		store:     store,
		nodeCache: make(map[kv.Key]*Node),
	}
	return ret
}

func (tr *Trie) Clone() *Trie {
	ret := &Trie{
		model:     tr.model,
		store:     tr.store,
		nodeCache: make(map[kv.Key]*Node),
	}
	for k, v := range tr.nodeCache {
		ret.nodeCache[k] = v.Clone()
	}
	return ret
}

func (tr *Trie) RootCommitment() VCommitment {
	n, ok := tr.GetNode(nil)
	if !ok {
		return nil
	}
	return tr.CommitToNode(n)
}

// GetNode takes node from the cache of fetches it from kv store
func (tr *Trie) GetNode(key []byte) (*Node, bool) {
	k := kv.Key(key)
	node, ok := tr.nodeCache[k]
	if ok {
		return node, true
	}
	nodeBin := tr.store.MustGet(k)
	if nodeBin == nil {
		return nil, false
	}
	node, err := NodeFromBytes(tr.model, nodeBin)
	assert(err == nil, err)

	tr.nodeCache[k] = node
	return node, true

}

func (tr *Trie) CommitToNode(n *Node) VCommitment {
	return tr.model.CommitToNode(n)
}

func (tr *Trie) ApplyMutations(store kv.KVWriter) {
	for k, v := range tr.nodeCache {
		if !v.IsEmpty() {
			store.Set(k, MustBytes(v))
		}
	}
	for k, v := range tr.nodeCache {
		if v.IsEmpty() {
			fmt.Printf("$$$$$$$$$$$$$$$$$$$$$$$ deleting '%s' with path fragment '%s'\n", k, string(v.PathFragment))
			store.Del(k)
		}
	}
}

func (tr *Trie) ClearCache() {
	tr.nodeCache = make(map[kv.Key]*Node)
}

// newTerminalNode assumes Key does not exist in the Trie
func (tr *Trie) newTerminalNode(key, pathFragment []byte, newTerminal TCommitment) *Node {
	ret := NewNode(pathFragment)
	ret.ModifiedTerminal = newTerminal
	tr.nodeCache[kv.Key(key)] = ret
	return ret
}

func (tr *Trie) newNodeCopy(key, pathFragment []byte, copyFrom *Node) *Node {
	ret := *copyFrom
	ret.PathFragment = pathFragment
	tr.nodeCache[kv.Key(key)] = &ret
	return &ret
}

// Delete deletes Key/value from the Trie
func (tr *Trie) Delete(key []byte) {
	tr.Update(key, nil)
}

// Commit calculates a new root commitment value from the cache and commits all mutations in the cached nodes
// Doesn't delete cached nodes
func (tr *Trie) Commit() {
	root, ok := tr.GetNode(nil)
	if !ok {
		return
	}
	tr.UpdateNodeCommitment(root)
}

func (tr *Trie) UpdateNodeCommitment(n *Node) VCommitment {
	if n == nil {
		// no node, no commitment
		return nil
	}
	n.Terminal = n.ModifiedTerminal
	for i, child := range n.ModifiedChildren {
		c := tr.UpdateNodeCommitment(child)
		if c != nil {
			if n.Children[i] == nil {
				n.Children[i] = tr.model.NewVectorCommitment()
			}
			n.Children[i].Update(c)
		} else {
			// deletion
			delete(n.Children, i)
		}
	}
	if len(n.ModifiedChildren) > 0 {
		n.ModifiedChildren = make(map[byte]*Node)
	}
	ret := tr.model.CommitToNode(n)
	assert((ret == nil) == n.IsEmpty(), "assert: (ret==nil) == n.IsEmpty()")
	return ret
}

// ProofGeneric returns generic proof path. Contains references trie node cache.
// Should be immediately converted into the specific proof model independent of the trie
// Normally only called by the model
func (tr *Trie) ProofGeneric(key []byte) *ProofGeneric {
	if len(key) == 0 {
		key = []byte{}
	}
	p, _, _, ending := tr.proofPath(key, 0)
	return &ProofGeneric{
		Key:    key,
		Path:   p,
		Ending: ending,
	}
}

// Update updates Trie with the key/value.
// value == nil means deletion
func (tr *Trie) Update(key []byte, value []byte) {
	c := tr.model.CommitToData(value)
	proof, lastKey, lastCommonPrefix, ending := tr.proofPath(key, 0)
	if len(proof) == 0 {
		if c != nil {
			tr.newTerminalNode(nil, key, c)
		}
		return

	}
	last := proof[len(proof)-1].Node
	switch ending {
	case EndingTerminal:
		// deleting means c == nil
		last.ModifiedTerminal = c

	case EndingExtend:
		if c == nil {
			// deleting: nothing change
			return
		}
		childIndexPosition := len(lastKey) + len(lastCommonPrefix)
		assert(childIndexPosition < len(key), "childPosition < len(key)")
		childIndex := key[childIndexPosition]
		assert(last.Children[childIndex] == nil, "last.Children[key[childPosition]] == nil")
		assert(last.ModifiedChildren[childIndex] == nil, "last.ModifiedChildren[key[childPosition]] == nil")
		last.ModifiedChildren[childIndex] = tr.newTerminalNode(key[:childIndexPosition+1], key[childIndexPosition+1:], c)

	case EndingSplit:
		if c == nil {
			// deleting: nothing change
			return
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
		insertNode := tr.newNodeCopy(keyContinue, last.PathFragment[splitChildIndex+1:], last)
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
			last.ModifiedChildren[childForkIndex] = tr.newTerminalNode(keyFork, key[len(keyFork):], c)
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
func (tr *Trie) proofPath(path []byte, pathPosition int) ([]ProofGenericElement, []byte, []byte, ProofEndingCode) {
	node, ok := tr.GetNode(nil)
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
		node, ok = tr.GetNode(key)
		if !ok {
			panic(fmt.Sprintf("inconsistency: trie key not found: '%s'", string(key)))
		}
		proof = append(proof, ProofGenericElement{Key: key, Node: node})
	}
}

func (tr *Trie) VectorCommitmentFromBytes(data []byte) (VCommitment, error) {
	ret := tr.model.NewVectorCommitment()
	if err := ret.Read(bytes.NewReader(data)); err != nil {
		return nil, err
	}
	return ret, nil
}

// Reconcile returns a list of keys in the store which cannot be proven in the trie
func (tr *Trie) Reconcile(store kv.KVMustIterator) []kv.Key {
	ret := make([]kv.Key, 0)
	store.MustIterate("", func(k kv.Key, v []byte) bool {
		p, _, _, ending := tr.proofPath([]byte(k), 0)
		if ending == EndingTerminal {
			if !EqualCommitments(tr.model.CommitToData(v), p[len(p)-1].Node.Terminal) {
				ret = append(ret, k)
			}
		} else {
			ret = append(ret, k)
		}
		return true
	})
	return ret
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

// ComputeCommitment computes commitment to arbitrary key/value iterator
func ComputeCommitment(model CommitmentModel, store kv.KVIterator) (VCommitment, error) {
	emptyStore := dict.New()
	tr := New(model, emptyStore)

	err := store.Iterate("", func(key kv.Key, value []byte) bool {
		tr.Update([]byte(key), value)
		return true
	})
	if err != nil {
		return nil, err
	}
	tr.Commit()
	return tr.RootCommitment(), nil
}
