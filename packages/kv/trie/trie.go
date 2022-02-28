package trie

import (
	"bytes"
	"fmt"
	"github.com/iotaledger/wasp/packages/kv"
	"golang.org/x/xerrors"
)

type Trie struct {
	access    accessTrie
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
		access:    *NewTrieAccess(store, model),
		nodeCache: make(map[kv.Key]*Node),
		deleted:   make(map[kv.Key]struct{}),
	}
	return ret
}

func (tr *Trie) Clone() *Trie {
	ret := &Trie{
		access:    tr.access,
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

// GetNode takes node from the cache of fetches it from kv store
func (tr *Trie) GetNode(key kv.Key) (*Node, bool) {
	if tr.isDeleted(key) {
		return nil, false
	}
	node, ok := tr.nodeCache[key]
	if ok {
		return node, true
	}
	node, ok = tr.access.GetNode(key)
	if !ok {
		return nil, false
	}
	tr.nodeCache[key] = node
	return node, true
}

func (tr *Trie) Model() CommitmentModel {
	return tr.access.model
}

func (tr *Trie) mustGetNode(key kv.Key) *Node {
	ret, ok := tr.GetNode(key)
	assert(ok, fmt.Sprintf("mustGetNode: not found key '%s'", key))
	return ret
}

func (tr *Trie) isDeleted(key kv.Key) bool {
	_, ret := tr.deleted[key]
	return ret
}

func (tr *Trie) removeKey(key kv.Key) {
	delete(tr.nodeCache, key)
	tr.deleted[key] = struct{}{}
}

func (tr *Trie) unDelete(key kv.Key) {
	delete(tr.deleted, key)
}

func (tr *Trie) CommitToNode(n *Node) VCommitment {
	return tr.access.model.CommitToNode(n)
}

func (tr *Trie) ApplyMutations(store kv.KVWriter) {
	for k, v := range tr.nodeCache {
		store.Set(k, v.Bytes())
	}
	for k := range tr.deleted {
		_, inCache := tr.nodeCache[k]
		assert(!inCache, "!inCache")
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
				n.ChildCommitments[childIndex] = tr.access.model.NewVectorCommitment()
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
	ret := tr.access.model.CommitToNode(n)
	return ret
}

// Update updates Trie with the key/value.
// value == nil means deletion
func (tr *Trie) Update(key []byte, value []byte) {
	c := tr.access.model.CommitToData(value)
	if c == nil {
		tr.Delete(key)
		return
	}

	proof, lastCommonPrefix, ending := proofPath(tr, key)
	if len(proof) == 0 {
		tr.newTerminalNode(nil, key, c)
		return
	}
	lastKey := proof[len(proof)-1]
	lastNode := tr.mustGetNode(kv.Key(lastKey))
	switch ending {
	case EndingTerminal:
		lastNode.NewTerminal = c

	case EndingExtend:
		childIndexPosition := len(lastKey) + len(lastCommonPrefix)
		assert(childIndexPosition < len(key), "childPosition < len(key)")
		childIndex := key[childIndexPosition]
		tr.removeKey(kv.Key(key[:childIndexPosition+1]))
		tr.newTerminalNode(key[:childIndexPosition+1], key[childIndexPosition+1:], c)
		lastNode.ModifiedChildren[childIndex] = struct{}{}

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
			assert(childForkIndex != childContinue, "childForkIndex != childContinue")
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
	proof, _, ending := proofPath(tr, key)
	if len(proof) == 0 || ending != EndingTerminal {
		return
	}
	lastKey := proof[len(proof)-1]
	lastNode, ok := tr.GetNode(kv.Key(lastKey))
	if !ok {
		return
	}
	lastNode.NewTerminal = nil
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

func (tr *Trie) markModifiedCommitmentsBackToRoot(proof [][]byte) {
	for i := len(proof) - 1; i > 0; i-- {
		k := proof[i]
		kPrev := proof[i-1]
		childIndex := k[len(k)-1]
		n := tr.mustGetNode(kv.Key(kPrev))
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

func (tr *Trie) checkReorg(key kv.Key, n *Node) (reorgStatus, byte) {
	if n.CommitsToTerminal() {
		return nodeReorgNOP, 0
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
		return nodeReorgRemove, 0
	case 1:
		for ret := range toCheck {
			return nodeReorgMerge, ret
		}
	}
	return nodeReorgNOP, 0
}

func (tr *Trie) mustCheckNode(key kv.Key) {
	n := tr.mustGetNode(key)
	status, _ := tr.checkReorg(key, n)
	assert(status == nodeReorgNOP, "status == nodeReorgNOP")
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
	ret := tr.access.model.NewVectorCommitment()
	if err := ret.Read(bytes.NewReader(data)); err != nil {
		return nil, err
	}
	return ret, nil
}

// Reconcile returns a list of keys in the store which cannot be proven in the trie
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
				if !EqualCommitments(tr.access.model.CommitToData(v), n.Terminal) {
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
