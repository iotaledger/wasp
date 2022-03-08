// Package trie implements functionality of generic verkle trie with 256 child commitment in each node
// + terminal commitment + commitment to the path fragment: 258 commitments in total.
// It mainly follows the definition from https://hackmd.io/@Evaldas/H13YFOVGt (except commitment to the path fragment)
// The commitment to the path fragment is needed to provide proofs of absence of keys
//
// The specific implementation of the commitment model is presented as a CommitmentModel interface
package trie

import (
	"bytes"
	"fmt"
	"github.com/iotaledger/wasp/packages/kv"
	"golang.org/x/xerrors"
)

// Trie is an updatable trie implemented on top of the key/value store. It is virtualized and optimized by caching of the
// trie update operation and keeping consistent trie in the cache
type Trie struct {
	// persisted trie
	nodeStore nodeStore
	// cached part of the trie
	nodeCache map[kv.Key]*Node
	// cached deleted nodes
	deleted map[kv.Key]struct{}
}

func New(model CommitmentModel, store kv.KVMustReader) *Trie {
	ret := &Trie{
		nodeStore: *NewNodeStore(store, model),
		nodeCache: make(map[kv.Key]*Node),
		deleted:   make(map[kv.Key]struct{}),
	}
	return ret
}

func (tr *Trie) Clone() *Trie {
	ret := &Trie{
		nodeStore: tr.nodeStore,
		nodeCache: make(map[kv.Key]*Node),
		deleted:   make(map[kv.Key]struct{}),
	}
	for k, v := range tr.nodeCache {
		ret.nodeCache[k] = v.Clone()
	}
	for k := range tr.deleted {
		ret.deleted[k] = struct{}{}
	}
	return ret
}

// Trie implements NodeStore interface. It caches all nodeStore for optimization purposes: multiple updates of trie do not require DB nodeStore
var _ NodeStore = &Trie{}

// GetNode takes node from the cache of fetches it from kv store
func (tr *Trie) GetNode(key kv.Key) (*Node, bool) {
	if _, isDeleted := tr.deleted[key]; isDeleted {
		return nil, false
	}
	node, ok := tr.nodeCache[key]
	if ok {
		return node, true
	}
	node, ok = tr.nodeStore.GetNode(key)
	if !ok {
		return nil, false
	}
	tr.nodeCache[key] = node
	return node, true
}

func (tr *Trie) Model() CommitmentModel {
	return tr.nodeStore.model
}

func (tr *Trie) mustGetNode(key kv.Key) *Node {
	ret, ok := tr.GetNode(key)
	assert(ok, fmt.Sprintf("mustGetNode: not found key '%s'", key))
	return ret
}

// removeKey marks key deleted
func (tr *Trie) removeKey(key kv.Key) {
	delete(tr.nodeCache, key)
	tr.deleted[key] = struct{}{}
}

// unDelete removes deletion mark, if any
func (tr *Trie) unDelete(key kv.Key) {
	delete(tr.deleted, key)
}

// CommitToNode calculates node commitment
func (tr *Trie) CommitToNode(n *Node) VCommitment {
	return tr.nodeStore.model.CommitToNode(n)
}

// PersistMutations persists the cache to the key/value store
// Does not clear cache
func (tr *Trie) PersistMutations(store kv.KVWriter) {
	for k, v := range tr.nodeCache {
		store.Set(k, v.Bytes())
	}
	for k := range tr.deleted {
		_, inCache := tr.nodeCache[k]
		assert(!inCache, "!inCache")
		store.Del(k)
	}
}

// ClearCache clears the node cache
func (tr *Trie) ClearCache() {
	tr.nodeCache = make(map[kv.Key]*Node)
	tr.deleted = make(map[kv.Key]struct{})
}

// newTerminalNode creates new node in the trie with specified pathFragment and terminal commitment.
// Assumes 'key' does not exist in the Trie
func (tr *Trie) newTerminalNode(key, pathFragment []byte, newTerminal TCommitment) *Node {
	tr.unDelete(kv.Key(key))
	ret := NewNode(pathFragment)
	ret.newTerminal = newTerminal
	_, already := tr.nodeCache[kv.Key(key)]
	assert(!already, "!already")
	tr.nodeCache[kv.Key(key)] = ret
	return ret
}

// newNodeCopy creates a new node by copying existing node (including cached part), assigning new path fragment and storing under new key
// Assumes noew with 'key' does not exist
func (tr *Trie) newNodeCopy(key, pathFragment []byte, copyFrom *Node) *Node {
	tr.unDelete(kv.Key(key))
	ret := *copyFrom
	ret.PathFragment = pathFragment
	tr.nodeCache[kv.Key(key)] = &ret
	return &ret
}

// Commit calculates a new root commitment value from the cache and commits all mutations in the cached nodeStore
// It is a re-calculation of the trie. Node caches are updated accordingly.
// Doesn't delete cached nodeStore
func (tr *Trie) Commit() {
	tr.UpdateNodeCommitment("")
}

// UpdateNodeCommitment re-calculates node commitment and, recursively, its children
// Child modification marks in 'modifiedChildren' are updated
func (tr *Trie) UpdateNodeCommitment(key kv.Key) VCommitment {
	n, ok := tr.GetNode(key)
	if !ok {
		// no node, no commitment
		return nil
	}
	n.Terminal = n.newTerminal
	for childIndex := range n.modifiedChildren {
		childKey := n.ChildKey(key, childIndex)
		c := tr.UpdateNodeCommitment(childKey)
		if c != nil {
			if n.ChildCommitments[childIndex] == nil {
				n.ChildCommitments[childIndex] = tr.nodeStore.model.NewVectorCommitment()
			}
			n.ChildCommitments[childIndex].Update(c)
		} else {
			// deletion
			delete(n.ChildCommitments, childIndex)
		}
	}
	if len(n.modifiedChildren) > 0 {
		n.modifiedChildren = make(map[byte]struct{})
	}
	ret := tr.nodeStore.model.CommitToNode(n)
	return ret
}

// Update updates Trie with the key/value. Reorganizes and re-calculates trie, keeps cache consistent
func (tr *Trie) Update(key []byte, value []byte) {
	c := tr.nodeStore.model.CommitToData(value)
	if c == nil {
		// nil value means deletion
		tr.Delete(key)
		return
	}
	// find path in the trie corresponding to the key
	proof, lastCommonPrefix, ending := proofPath(tr, key)
	if len(proof) == 0 {
		tr.newTerminalNode(nil, key, c)
		return
	}
	lastKey := proof[len(proof)-1]
	lastNode := tr.mustGetNode(kv.Key(lastKey))
	switch ending {
	case EndingTerminal:
		lastNode.newTerminal = c

	case EndingExtend:
		childIndexPosition := len(lastKey) + len(lastCommonPrefix)
		assert(childIndexPosition < len(key), "childPosition < len(key)")
		childIndex := key[childIndexPosition]
		tr.removeKey(kv.Key(key[:childIndexPosition+1]))
		tr.newTerminalNode(key[:childIndexPosition+1], key[childIndexPosition+1:], c)
		lastNode.modifiedChildren[childIndex] = struct{}{}

	case EndingSplit:
		childPosition := len(lastKey) + len(lastCommonPrefix)
		assert(childPosition <= len(key), "childPosition < len(key)")
		keyContinue := make([]byte, childPosition+1)
		copy(keyContinue, key)
		splitIndex := len(lastCommonPrefix)
		assert(splitIndex < len(lastNode.PathFragment), "splitIndex<len(last.Node.PathFragment)")
		childContinue := lastNode.PathFragment[splitIndex]
		keyContinue[len(keyContinue)-1] = childContinue

		// create new node on keyContinue, move everything from old to the new node and adjust the path fragment
		tr.newNodeCopy(keyContinue, lastNode.PathFragment[splitIndex+1:], lastNode)
		// clear the old one and adjust path fragment. Continue with 1 child, the new node
		lastNode.ChildCommitments = make(map[uint8]VCommitment)
		lastNode.modifiedChildren = make(map[uint8]struct{})
		lastNode.PathFragment = lastCommonPrefix
		lastNode.modifiedChildren[childContinue] = struct{}{}
		lastNode.Terminal = nil
		lastNode.newTerminal = nil
		// insert terminal
		if childPosition == len(key) {
			// no need for the new node
			lastNode.newTerminal = c
		} else {
			// create a new node
			keyFork := key[:len(keyContinue)]
			childForkIndex := keyFork[len(keyFork)-1]
			assert(childForkIndex != childContinue, "childForkIndex != childContinue")
			tr.newTerminalNode(keyFork, key[len(keyFork):], c)
			lastNode.modifiedChildren[childForkIndex] = struct{}{}
		}

	default:
		panic("inconsistency: unknown path ending code")
	}
	tr.markModifiedCommitmentsBackToRoot(proof)
}

// Delete deletes Key/value from the Trie, reorganizes the trie
func (tr *Trie) Delete(key []byte) {
	proof, _, ending := proofPath(tr, key)
	if len(proof) == 0 || ending != EndingTerminal {
		return
	}
	lastKey := proof[len(proof)-1]
	lastNode, ok := tr.GetNode(kv.Key(lastKey))
	if !ok {
		return
	}
	lastNode.newTerminal = nil
	reorg, mergeChildIndex := tr.checkReorg(kv.Key(lastKey), lastNode)
	switch reorg {
	case nodeReorgNOP:
		// do nothing
		tr.markModifiedCommitmentsBackToRoot(proof)
	case nodeReorgRemove:
		// last node does not commit to anything, should be removed
		tr.removeKey(kv.Key(lastKey))
		if len(proof) >= 2 {
			// check if
			prevKey := proof[len(proof)-2]
			prevNode, ok := tr.GetNode(kv.Key(prevKey))
			if !ok {
				panic(xerrors.Errorf("Delete: can't find node key '%s'", kv.Key(prevKey)))
			}
			tr.markModifiedCommitmentsBackToRoot(proof)
			reorg, mergeChildIndex = tr.checkReorg(kv.Key(prevKey), prevNode)
			if reorg == nodeReorgMerge {
				tr.mergeNode(prevKey, prevNode, mergeChildIndex)
			}
		}
	case nodeReorgMerge:
		tr.mergeNode(lastKey, lastNode, mergeChildIndex)
		tr.markModifiedCommitmentsBackToRoot(proof)
	}
}

// mergeNode merges nodes when it is possible, i.e. first node does not contain terminal commitment and has only one
// child commitment. In this case pathFragments can be merged in one resulting node
func (tr *Trie) mergeNode(key []byte, n *Node, childIndex byte) {
	nextKey := n.ChildKey(kv.Key(key), childIndex)
	nextNode := tr.mustGetNode(nextKey)
	var newPathFragment bytes.Buffer
	newPathFragment.Write(n.PathFragment)
	newPathFragment.WriteByte(childIndex)
	newPathFragment.Write(nextNode.PathFragment)

	tr.newNodeCopy(key, newPathFragment.Bytes(), nextNode)
	tr.removeKey(nextKey)
}

// markModifiedCommitmentsBackToRoot updates 'modifiedChildren' marks along tha path from the updated node to the root
func (tr *Trie) markModifiedCommitmentsBackToRoot(proof [][]byte) {
	for i := len(proof) - 1; i > 0; i-- {
		k := proof[i]
		kPrev := proof[i-1]
		childIndex := k[len(k)-1]
		n := tr.mustGetNode(kv.Key(kPrev))
		n.modifiedChildren[childIndex] = struct{}{}
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
	for childIndex := range n.modifiedChildren {
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

// checkReorg check what has to be done with the node after deletion: either nothing, node must be removed or merged
func (tr *Trie) checkReorg(key kv.Key, n *Node) (reorgStatus, byte) {
	if n.CommitsToTerminal() {
		return nodeReorgNOP, 0
	}
	toCheck := make(map[byte]struct{})
	for c := range n.ChildCommitments {
		toCheck[c] = struct{}{}
	}
	for c := range n.modifiedChildren {
		if tr.hasCommitment(n.ChildKey(key, c)) {
			toCheck[c] = struct{}{}
		} else {
			delete(toCheck, c)
		}
	}
	switch len(toCheck) {
	case 0:
		return nodeReorgRemove, 0
	case 1:
		for ret := range toCheck {
			return nodeReorgMerge, ret
		}
	}
	return nodeReorgNOP, 0
}

// UpdateStr updates key/value pair in the trie
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

// DeleteStr removes node from trie
func (tr *Trie) DeleteStr(key interface{}) {
	var k []byte
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
	tr.Delete(k)
}

func (tr *Trie) VectorCommitmentFromBytes(data []byte) (VCommitment, error) {
	ret := tr.nodeStore.model.NewVectorCommitment()
	if err := ret.Read(bytes.NewReader(data)); err != nil {
		return nil, err
	}
	return ret, nil
}

// Reconcile returns a list of keys in the store which cannot be proven in the trie
// Trie is consistent if empty slice is returned
// May be an expensive operation
func (tr *Trie) Reconcile(store kv.KVMustIterator) []kv.Key {
	ret := make([]kv.Key, 0)
	store.MustIterate("", func(k kv.Key, v []byte) bool {
		p, _, ending := proofPath(tr, []byte(k))
		if ending == EndingTerminal {
			lastKey := p[len(p)-1]
			n, ok := tr.GetNode(kv.Key(lastKey))
			if !ok {
				ret = append(ret, k)
			} else {
				if !EqualCommitments(tr.nodeStore.model.CommitToData(v), n.Terminal) {
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

// UpdateAll mass-updates trie from the key/value store.
// To be used to build trie for arbitrary key/value data sets
func (tr *Trie) UpdateAll(store kv.KVMustIterator) {
	store.MustIterate("", func(k kv.Key, v []byte) bool {
		tr.Update([]byte(k), v)
		return true
	})
}
