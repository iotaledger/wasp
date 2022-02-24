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
	deleted   map[kv.Key]struct{}
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
	n, ok := tr.GetNode("")
	if !ok {
		return nil
	}
	return tr.CommitToNode(n)
}

// GetNode takes node from the cache of fetches it from kv store
func (tr *Trie) GetNode(key kv.Key) (*Node, bool) {
	if _, deleted := tr.deleted[key]; deleted {
		return nil, false
	}
	node, ok := tr.nodeCache[key]
	if ok {
		return node, true
	}
	nodeBin := tr.store.MustGet(key)
	if nodeBin == nil {
		return nil, false
	}
	node, err := NodeFromBytes(tr.model, nodeBin)
	assert(err == nil, err)

	tr.nodeCache[key] = node
	return node, true

}

func (tr *Trie) CommitToNode(n *Node) VCommitment {
	return tr.model.CommitToNode(n)
}

func (tr *Trie) ApplyMutations(store kv.KVWriter) {
	for k, v := range tr.nodeCache {
		store.Set(k, MustBytes(v))
	}
	for k := range tr.deleted {
		_, inCache := tr.nodeCache[k]
		assert(!inCache, "!inCache")
		fmt.Printf("$$$$$$$$$$$$$$$$$$$$$$$ deleting key '%s' from trie", k)
		store.Del(k)
	}
}

func (tr *Trie) ClearCache() {
	tr.nodeCache = make(map[kv.Key]*Node)
	tr.deleted = make(map[kv.Key]struct{})
}

// newTerminalNode assumes Key does not exist in the Trie
func (tr *Trie) newTerminalNode(key, pathFragment []byte, newTerminal TCommitment) *Node {
	ret := NewNode(pathFragment)
	ret.NewTerminal = newTerminal
	_, already := tr.nodeCache[kv.Key(key)]
	assert(!already, "!already")
	tr.nodeCache[kv.Key(key)] = ret
	return ret
}

func (tr *Trie) newNodeCopy(key, pathFragment []byte, copyFrom *Node) *Node {
	ret := *copyFrom
	ret.PathFragment = pathFragment
	tr.nodeCache[kv.Key(key)] = &ret
	return &ret
}

// Commit calculates a new root commitment value from the cache and commits all mutations in the cached nodes
// Doesn't delete cached nodes
func (tr *Trie) Commit() {
	tr.UpdateNodeCommitment("")
}

func (tr *Trie) UpdateNodeCommitment(key kv.Key) VCommitment {
	n, ok := tr.GetNode(key)
	if !ok {
		// no node, no commitment
		return nil
	}
	n.Terminal = n.NewTerminal
	for childIndex := range n.ModifiedChildren {
		childKey := n.ChildKey(key, childIndex)
		c := tr.UpdateNodeCommitment(childKey)
		if c != nil {
			if n.ChildCommitments[childIndex] == nil {
				n.ChildCommitments[childIndex] = tr.model.NewVectorCommitment()
			}
			n.ChildCommitments[childIndex].Update(c)
		} else {
			// deletion
			delete(n.ChildCommitments, childIndex)
		}
	}
	if len(n.ModifiedChildren) > 0 {
		n.ModifiedChildren = make(map[byte]struct{})
	}
	ret := tr.model.CommitToNode(n)
	return ret
}

// ProofGeneric returns generic proof path. Contains references trie node cache.
// Should be immediately converted into the specific proof model independent of the trie
// Normally only called by the model
func (tr *Trie) ProofGeneric(key []byte) *ProofGeneric {
	if len(key) == 0 {
		key = []byte{}
	}
	p, _, _, ending := tr.proofPath(key)
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
	if c == nil {
		tr.Delete(key)
		return
	}

	proof, lastKey, lastCommonPrefix, ending := tr.proofPath(key)
	if len(proof) == 0 {
		tr.newTerminalNode(nil, key, c)
		return
	}
	last := proof[len(proof)-1].Node
	switch ending {
	case EndingTerminal:
		last.NewTerminal = c

	case EndingExtend:
		childIndexPosition := len(lastKey) + len(lastCommonPrefix)
		assert(childIndexPosition < len(key), "childPosition < len(key)")
		childIndex := key[childIndexPosition]
		tr.newTerminalNode(key[:childIndexPosition+1], key[childIndexPosition+1:], c)
		last.ModifiedChildren[childIndex] = struct{}{}

	case EndingSplit:
		childPosition := len(lastKey) + len(lastCommonPrefix)
		assert(childPosition <= len(key), "childPosition < len(key)")
		keyContinue := make([]byte, childPosition+1)
		copy(keyContinue, key)
		splitChildIndex := len(lastCommonPrefix)
		assert(splitChildIndex < len(last.PathFragment), "splitChildIndex<len(last.Node.PathFragment)")
		childContinue := last.PathFragment[splitChildIndex]
		keyContinue[len(keyContinue)-1] = childContinue

		// create new node on keyContinue, move everything from old to the new node and adjust the path fragment
		tr.newNodeCopy(keyContinue, last.PathFragment[splitChildIndex+1:], last)
		// clear the old one and adjust path fragment. Continue with 1 child, the new node
		last.ChildCommitments = make(map[uint8]VCommitment)
		last.ModifiedChildren = make(map[uint8]struct{})
		last.PathFragment = lastCommonPrefix
		last.ModifiedChildren[childContinue] = struct{}{}
		last.Terminal = nil
		last.NewTerminal = nil
		// insert terminal
		if childPosition == len(key) {
			// no need for the new node
			last.NewTerminal = c
		} else {
			// create a new node
			keyFork := key[:len(keyContinue)]
			childForkIndex := keyFork[len(keyFork)-1]
			assert(int(childForkIndex) != splitChildIndex, "childForkIndex != splitChildIndex")
			tr.newTerminalNode(keyFork, key[len(keyFork):], c)
			last.ModifiedChildren[childForkIndex] = struct{}{}
		}

	default:
		panic("inconsistency: unknown path ending code")
	}
	// update commitment path back to root
	for i := len(proof) - 2; i >= 0; i-- {
		k := proof[i+1].Key
		childIndex := k[len(k)-1]
		proof[i].Node.ModifiedChildren[childIndex] = struct{}{}
	}
}

func (tr *Trie) HasCommitment(key kv.Key) bool {
	n, ok := tr.GetNode(key)
	if !ok {
		return false
	}
	if n.NewTerminal != nil {
		return true
	}
	if n.Terminal == nil {
		return false
	}
	for childIndex := range n.ModifiedChildren {
		if tr.HasCommitment(n.ChildKey(key, childIndex)) {
			return true
		}
	}
	if len(n.ChildCommitments) > 0 {
		return true
	}
	return false
}

// OneChildCommitted checks condition of the node to be merged
func (tr *Trie) OneChildCommitted(key kv.Key) (byte, bool) {
	n, ok := tr.GetNode(key)
	if !ok {
		return 0, false
	}
	if n.NewTerminal != nil || n.Terminal == nil {
		return 0, false
	}
	coms := make(map[byte]struct{})
	for c := range n.ChildCommitments {
		coms[c] = struct{}{}
	}
	for c := range n.ModifiedChildren {
		if tr.HasCommitment(n.ChildKey(key, c)) {
			coms[c] = struct{}{}
		}
	}
	if len(coms) != 1 {
		return 0, false
	}
	var c byte
	for tmp := range coms {
		c = tmp
		break
	}
	return c, true
}

// Delete deletes Key/value from the Trie
func (tr *Trie) Delete(key []byte) {
	proof, _, _, ending := tr.proofPath(key)
	if len(proof) == 0 {
		return
	}
	if ending != EndingTerminal {
		return
	}
	last := proof[len(proof)-1]
	last.Node.NewTerminal = nil
	// if it is empty, mark it as deleted
	if !tr.HasCommitment(kv.Key(last.Key)) {
		tr.deleted[kv.Key(key)] = struct{}{}

		for i := len(proof) - 2; i >= 0; i-- {
			k := proof[i+1].Key
			childIndex := k[len(k)-1]
			proof[i].Node.ModifiedChildren[childIndex] = struct{}{}
		}
		return
	}
	// if only one committed child left
	childIndex, ok := tr.OneChildCommitted(kv.Key(last.Key))
	if !ok {
		return
	}
	// TODO merge nodes

}

func (tr *Trie) UpdateStr(key interface{}, value interface{}) {
	var k, v []byte
	if key != nil {
		switch kt := key.(type) {
		case []byte:
			k = kt
		case string:
			k = []byte(kt)
		default:
			panic("[]byte or string expected")
		}
	}
	if value != nil {
		switch vt := value.(type) {
		case []byte:
			v = vt
		case string:
			v = []byte(vt)
		default:
			panic("[]byte or string expected")
		}
	}
	tr.Update(k, v)
}

func (tr *Trie) DeleteStr(key interface{}) {
	tr.UpdateStr(key, nil)
}

// returns key of the last node and common prefix with the fragment
func (tr *Trie) proofPath(path []byte) ([]ProofGenericElement, []byte, []byte, ProofEndingCode) {
	node, ok := tr.GetNode("")
	if !ok {
		return nil, nil, nil, 0
	}

	proof := make([]ProofGenericElement, 0)
	var key []byte

	for {
		// it means we continue the branch of commitment
		proof = append(proof, ProofGenericElement{Key: key, Node: node})
		assert(len(path) <= len(key), "pathPosition<=len(path)")
		if bytes.Equal(path[len(key):], node.PathFragment) {
			return proof, nil, nil, EndingTerminal
		}
		prefix := commonPrefix(path[len(key):], node.PathFragment)

		if len(prefix) < len(node.PathFragment) {
			return proof, key, prefix, EndingSplit
		}
		assert(len(prefix) == len(node.PathFragment), "len(prefix)==len(node.PathFragment)")
		childIndexPosition := len(key) + len(prefix)
		assert(childIndexPosition < len(path), "childIndexPosition<len(path)")

		childKey := node.ChildKey(kv.Key(key), path[childIndexPosition])

		node, ok = tr.GetNode(childKey)
		if !ok {
			// if there are no commitment to the child at the position, it means trie must be extended at this point
			return proof, key, prefix, EndingExtend
		}
		key = []byte(childKey)
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
		p, _, _, ending := tr.proofPath([]byte(k))
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
