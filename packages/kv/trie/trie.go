package trie

import (
	"bytes"
	"fmt"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"golang.org/x/xerrors"
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
	Path   [][]byte
	Ending ProofEndingCode
}

type ProofGenericElement struct {
	Key []byte
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
	if tr.isDeleted(key) {
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

func (tr *Trie) isDeleted(key kv.Key) bool {
	_, ret := tr.deleted[key]
	return ret
}

func (tr *Trie) markDeleted(key kv.Key) {
	tr.deleted[key] = struct{}{}
}

func (tr *Trie) unDelete(key kv.Key) {
	delete(tr.deleted, key)
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
	tr.unDelete(kv.Key(key))
	ret := NewNode(pathFragment)
	ret.NewTerminal = newTerminal
	_, already := tr.nodeCache[kv.Key(key)]
	assert(!already, "!already")
	tr.nodeCache[kv.Key(key)] = ret
	return ret
}

func (tr *Trie) newNodeCopy(key, pathFragment []byte, copyFrom *Node) *Node {
	tr.unDelete(kv.Key(key))
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
	p, _, ending := tr.proofPath(key)
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

	proof, lastCommonPrefix, ending := tr.proofPath(key)
	if len(proof) == 0 {
		tr.newTerminalNode(nil, key, c)
		return
	}
	lastKey := proof[len(proof)-1]
	lastNode, ok := tr.GetNode(kv.Key(lastKey))
	if !ok {
		panic(xerrors.Errorf("Update: can't find node key '%s'", kv.Key(lastKey)))
	}
	switch ending {
	case EndingTerminal:
		lastNode.NewTerminal = c

	case EndingExtend:
		childIndexPosition := len(lastKey) + len(lastCommonPrefix)
		assert(childIndexPosition < len(key), "childPosition < len(key)")
		childIndex := key[childIndexPosition]
		tr.newTerminalNode(key[:childIndexPosition+1], key[childIndexPosition+1:], c)
		lastNode.ModifiedChildren[childIndex] = struct{}{}

	case EndingSplit:
		childPosition := len(lastKey) + len(lastCommonPrefix)
		assert(childPosition <= len(key), "childPosition < len(key)")
		keyContinue := make([]byte, childPosition+1)
		copy(keyContinue, key)
		splitChildIndex := len(lastCommonPrefix)
		assert(splitChildIndex < len(lastNode.PathFragment), "splitChildIndex<len(last.Node.PathFragment)")
		childContinue := lastNode.PathFragment[splitChildIndex]
		keyContinue[len(keyContinue)-1] = childContinue

		// create new node on keyContinue, move everything from old to the new node and adjust the path fragment
		tr.newNodeCopy(keyContinue, lastNode.PathFragment[splitChildIndex+1:], lastNode)
		// clear the old one and adjust path fragment. Continue with 1 child, the new node
		lastNode.ChildCommitments = make(map[uint8]VCommitment)
		lastNode.ModifiedChildren = make(map[uint8]struct{})
		lastNode.PathFragment = lastCommonPrefix
		lastNode.ModifiedChildren[childContinue] = struct{}{}
		lastNode.Terminal = nil
		lastNode.NewTerminal = nil
		// insert terminal
		if childPosition == len(key) {
			// no need for the new node
			lastNode.NewTerminal = c
		} else {
			// create a new node
			keyFork := key[:len(keyContinue)]
			childForkIndex := keyFork[len(keyFork)-1]
			assert(int(childForkIndex) != splitChildIndex, "childForkIndex != splitChildIndex")
			tr.newTerminalNode(keyFork, key[len(keyFork):], c)
			lastNode.ModifiedChildren[childForkIndex] = struct{}{}
		}

	default:
		panic("inconsistency: unknown path ending code")
	}
	tr.markModifiedCommitmentsBackToRoot(proof)
}

// Delete deletes Key/value from the Trie
func (tr *Trie) Delete(key []byte) {
	proof, _, ending := tr.proofPath(key)
	if len(proof) == 0 || ending != EndingTerminal {
		return
	}
	lastKey := proof[len(proof)-1]
	lastNode, reorg, mergeChildIndex := tr.removeTerminal(kv.Key(lastKey))
	if lastNode == nil {
		return
	}
	switch reorg {
	case nodeReorgNOP:
		// do nothing
	case nodeReorgRemove:
		// last node does not commit to anything, should be removed
		tr.markDeleted(kv.Key(lastKey))
	case nodeReorgMerge:
		// last node commits to exactly one child, must be merged at mergeChildIndex
		nextKey := lastNode.ChildKey(kv.Key(lastKey), mergeChildIndex)
		nextNode, ok := tr.GetNode(nextKey)
		if !ok {
			panic(xerrors.Errorf("Delete: can't find child node key '%s'", kv.Key(nextKey)))
		}
		newPathFragment := make([]byte, 0, len(lastNode.PathFragment)+len(nextNode.PathFragment))
		newPathFragment = append(newPathFragment, lastNode.PathFragment...)
		newPathFragment = append(newPathFragment, nextNode.PathFragment...)

		tr.newNodeCopy(lastKey, newPathFragment, nextNode)
		tr.markDeleted(nextKey)
	}
	tr.markModifiedCommitmentsBackToRoot(proof)
}

func (tr *Trie) markModifiedCommitmentsBackToRoot(proof [][]byte) {
	for i := len(proof) - 1; i > 0; i-- {
		k := proof[i]
		kPrev := proof[i-1]
		childIndex := k[len(k)-1]
		n, ok := tr.GetNode(kv.Key(kPrev))
		if !ok {
			panic(xerrors.Errorf("markModifiedCommitmentsBackToRoot: can't find node key '%s'", kv.Key(kPrev)))
		}
		n.ModifiedChildren[childIndex] = struct{}{}
	}
}

// hasCommitment returns if trie will contain commitment to the key in the (future) committed state
func (tr *Trie) hasCommitment(key kv.Key) bool {
	n, ok := tr.GetNode(key)
	if !ok {
		return false
	}
	if n.CommitsToTerminal() {
		// commits to terminal
		return true
	}
	for childIndex := range n.ModifiedChildren {
		if tr.hasCommitment(n.ChildKey(key, childIndex)) {
			// modified child commits to something
			return true
		}
	}
	// new commitments do not come from children
	if len(n.ChildCommitments) > 0 {
		// existing children commit
		return true
	}
	// node does not commit to anything
	return false
}

type reorgStatus int

const (
	nodeReorgRemove = reorgStatus(iota)
	nodeReorgMerge
	nodeReorgNOP
)

func (tr *Trie) removeTerminal(key kv.Key) (*Node, reorgStatus, byte) {
	n, ok := tr.GetNode(key)
	if !ok {
		return nil, nodeReorgNOP, 0
	}
	n.NewTerminal = nil
	if n.CommitsToTerminal() {
		return n, nodeReorgNOP, 0
	}
	toCheck := make(map[byte]struct{})
	for c := range n.ChildCommitments {
		toCheck[c] = struct{}{}
	}
	for c := range n.ModifiedChildren {
		if tr.hasCommitment(n.ChildKey(key, c)) {
			toCheck[c] = struct{}{}
		} else {
			delete(toCheck, c)
		}
	}
	switch len(toCheck) {
	case 0:
		return n, nodeReorgRemove, 0
	case 1:
		for ret := range toCheck {
			return n, nodeReorgMerge, ret
		}
	}
	return n, nodeReorgNOP, 0
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

// returns
// - path of keys which leads to 'path'
// - common prefix between the last key and the fragment
func (tr *Trie) proofPath(path []byte) ([][]byte, []byte, ProofEndingCode) {
	node, ok := tr.GetNode("")
	if !ok {
		return nil, nil, 0
	}

	proof := make([][]byte, 0)
	var key []byte

	for {
		// it means we continue the branch of commitment
		proof = append(proof, key)
		assert(len(path) <= len(key), "pathPosition<=len(path)")
		if bytes.Equal(path[len(key):], node.PathFragment) {
			return proof, nil, EndingTerminal
		}
		prefix := commonPrefix(path[len(key):], node.PathFragment)

		if len(prefix) < len(node.PathFragment) {
			return proof, prefix, EndingSplit
		}
		assert(len(prefix) == len(node.PathFragment), "len(prefix)==len(node.PathFragment)")
		childIndexPosition := len(key) + len(prefix)
		assert(childIndexPosition < len(path), "childIndexPosition<len(path)")

		childKey := node.ChildKey(kv.Key(key), path[childIndexPosition])

		node, ok = tr.GetNode(childKey)
		if !ok {
			// if there are no commitment to the child at the position, it means trie must be extended at this point
			return proof, prefix, EndingExtend
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
		p, _, ending := tr.proofPath([]byte(k))
		if ending == EndingTerminal {
			lastKey := p[len(p)-1]
			n, ok := tr.GetNode(kv.Key(lastKey))
			if !ok {
				ret = append(ret, k)
			} else {
				if !EqualCommitments(tr.model.CommitToData(v), n.Terminal) {
					ret = append(ret, k)
				}
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
