package trie

import (
	"bytes"
	"encoding/hex"
)

// bufferedNode is a modified node
type bufferedNode struct {
	// persistent
	nodeData            *nodeData
	value               []byte // will be persisted in value store if not nil
	terminal            *tcommitment
	pathExtension       []byte
	uncommittedChildren map[byte]*bufferedNode // children which has been modified
	triePath            []byte
}

func newBufferedNode(n *nodeData, triePath []byte) *bufferedNode {
	if n == nil {
		n = newNodeData()
	}
	ret := &bufferedNode{
		nodeData:            n,
		terminal:            n.terminal,
		pathExtension:       n.pathExtension,
		uncommittedChildren: make(map[byte]*bufferedNode),
		triePath:            triePath,
	}
	return ret
}

// commitNode re-calculates node commitment and, recursively, its children commitments
func (n *bufferedNode) commitNode(triePartition, valuePartition KVWriter) {
	childUpdates := make(map[byte]*Hash)
	for idx, child := range n.uncommittedChildren {
		if child == nil {
			childUpdates[idx] = nil
		} else {
			child.commitNode(triePartition, valuePartition)
			childUpdates[idx] = &child.nodeData.commitment
		}
	}
	n.nodeData.update(childUpdates, n.terminal, n.pathExtension)

	n.mustPersist(triePartition)
	if len(n.value) > 0 {
		valuePartition.Set(n.terminal.Bytes(), n.value)
	}
}

func (n *bufferedNode) mustPersist(w KVWriter) {
	dbKey := n.nodeData.commitment.Bytes()
	var buf bytes.Buffer
	err := n.nodeData.Write(&buf)
	assertNoError(err)
	w.Set(dbKey, buf.Bytes())
}

func (n *bufferedNode) isRoot() bool {
	return len(n.triePath) == 0
}

// indexAsChild return index of the node as a child in the parent commitment and flag if it is a mutatedRoot
func (n *bufferedNode) indexAsChild() byte {
	assertf(!n.isRoot(), "indexAsChild:: receiver can't be a root node")
	return n.triePath[len(n.triePath)-1]
}

//nolint:unparam // for later use of idx
func (n *bufferedNode) setModifiedChild(child *bufferedNode, idx ...byte) {
	var index byte

	if child != nil {
		index = child.indexAsChild()
	} else {
		assertf(len(idx) > 0, "setModifiedChild: index of the child must be specified if the child is nil")
		index = idx[0]
	}
	n.uncommittedChildren[index] = child
}

func (n *bufferedNode) removeChild(child *bufferedNode, idx ...byte) {
	var index byte
	if child == nil {
		assertf(len(idx) > 0, "child index must be specified")
		index = idx[0]
	} else {
		index = child.indexAsChild()
	}
	n.uncommittedChildren[index] = nil
}

func (n *bufferedNode) setPathExtension(pf []byte) {
	n.pathExtension = pf
}

func (n *bufferedNode) setValue(value []byte) {
	if len(value) == 0 {
		n.terminal = nil
		n.value = nil
		return
	}
	n.terminal = commitToData(value)
	_, valueIsInCommitment := n.terminal.ExtractValue()
	if valueIsInCommitment {
		n.value = nil
	} else {
		n.value = value
	}
}

func (n *bufferedNode) setTriePath(triePath []byte) {
	n.triePath = triePath
}

func (n *bufferedNode) getChild(childIndex byte, db *nodeStore) *bufferedNode {
	if ret, already := n.uncommittedChildren[childIndex]; already {
		return ret
	}
	childCommitment := n.nodeData.children[childIndex]
	if childCommitment == nil {
		return nil
	}
	childTriePath := concat(n.triePath, n.pathExtension, []byte{childIndex})

	nodeFetched, ok := db.FetchNodeData(*childCommitment)
	assertf(ok, "TrieUpdatable::getChild: can't fetch node. triePath: '%s', dbKey: '%s",
		hex.EncodeToString(childCommitment.Bytes()), hex.EncodeToString(childTriePath))

	return newBufferedNode(nodeFetched, childTriePath)
}

// node is in the trie if at least one of the two is true:
// - it commits to terminal
// - it commits to at least 2 children
// Otherwise node has to be merged/removed
// It can only happen during deletion
func (n *bufferedNode) hasToBeRemoved(nodeStore *nodeStore) (bool, *bufferedNode) {
	if n.terminal != nil {
		return false, nil
	}
	var theOnlyChildCommitted *bufferedNode

	for i := 0; i < NumChildren; i++ {
		child := n.getChild(byte(i), nodeStore)
		if child != nil {
			if theOnlyChildCommitted != nil {
				// at least 2 children
				return false, nil
			}
			theOnlyChildCommitted = child
		}
	}
	return true, theOnlyChildCommitted
}
