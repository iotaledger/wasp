package trie

import (
	"encoding/hex"
)

// draftNode is a modified node
type draftNode struct {
	// persistent
	nodeData            *NodeData
	value               []byte // will be persisted in value store if not nil
	terminal            *Tcommitment
	pathExtension       []byte
	uncommittedChildren map[byte]*draftNode // children which has been modified
	triePath            []byte
}

func newDraftNode(n *NodeData, triePath []byte) *draftNode {
	if n == nil {
		n = newNodeData()
	}
	ret := &draftNode{
		nodeData:            n,
		terminal:            n.Terminal,
		pathExtension:       n.PathExtension,
		uncommittedChildren: make(map[byte]*draftNode),
		triePath:            triePath,
	}
	return ret
}

func (n *draftNode) CommitsToExternalValue() bool {
	return n.terminal != nil && !n.terminal.IsValue
}

func (n *draftNode) isRoot() bool {
	return len(n.triePath) == 0
}

// indexAsChild return index of the node as a child in the parent commitment and flag if it is a mutatedRoot
func (n *draftNode) indexAsChild() byte {
	assertf(!n.isRoot(), "indexAsChild:: receiver can't be a root node")
	return n.triePath[len(n.triePath)-1]
}

//nolint:unparam // for later use of idx
func (n *draftNode) setModifiedChild(child *draftNode, idx ...byte) {
	var index byte

	if child != nil {
		index = child.indexAsChild()
	} else {
		assertf(len(idx) > 0, "setModifiedChild: index of the child must be specified if the child is nil")
		index = idx[0]
	}
	n.uncommittedChildren[index] = child
}

func (n *draftNode) removeChild(child *draftNode, idx ...byte) {
	var index byte
	if child == nil {
		assertf(len(idx) > 0, "child index must be specified")
		index = idx[0]
	} else {
		index = child.indexAsChild()
	}
	n.uncommittedChildren[index] = nil
}

func (n *draftNode) setPathExtension(pf []byte) {
	n.pathExtension = pf
}

func (n *draftNode) setValue(value []byte) {
	if len(value) == 0 {
		n.terminal = nil
		n.value = nil
		return
	}
	n.terminal = CommitToData(value)
	_, valueIsInCommitment := n.terminal.ExtractValue()
	if valueIsInCommitment {
		n.value = nil
	} else {
		n.value = value
	}
}

func (n *draftNode) setTriePath(triePath []byte) {
	n.triePath = triePath
}

func (n *draftNode) getChild(childIndex byte, tr *TrieR) *draftNode {
	if ret, already := n.uncommittedChildren[childIndex]; already {
		return ret
	}
	childCommitment := n.nodeData.Children[childIndex]
	if childCommitment == nil {
		return nil
	}
	childTriePath := concat(n.triePath, n.pathExtension, []byte{childIndex})

	nodeFetched, ok := tr.fetchNodeData(*childCommitment)
	assertf(ok, "TrieDraft::getChild: can't fetch node. triePath: '%s', dbKey: '%s",
		hex.EncodeToString(childCommitment.Bytes()), hex.EncodeToString(childTriePath))

	return newDraftNode(nodeFetched, childTriePath)
}

// node is in the trie if at least one of the two is true:
// - it commits to terminal
// - it commits to at least 2 children
// Otherwise node has to be merged/removed
// It can only happen during deletion
func (n *draftNode) hasToBeRemoved(tr *TrieR) (bool, *draftNode) {
	if n.terminal != nil {
		return false, nil
	}
	var theOnlyChildCommitted *draftNode

	for i := range byte(NumChildren) {
		child := n.getChild(i, tr)
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

// traversePreOrder traverses the modified nodes pre-order
func (n *draftNode) traversePreOrder(f func(*draftNode) IterateNodesAction) bool {
	action := f(n)
	if action == IterateStop {
		return false
	}
	if action != IterateSkipSubtree {
		for i := range byte(NumChildren) {
			if child := n.uncommittedChildren[i]; child != nil {
				if !child.traversePreOrder(f) {
					return false
				}
			}
		}
	}
	return true
}

// traversePostOrder traverses the modified nodes post-order
func (n *draftNode) traversePostOrder(f func(*draftNode)) {
	for i := range byte(NumChildren) {
		if child := n.uncommittedChildren[i]; child != nil {
			child.traversePostOrder(f)
		}
	}
	f(n)
}
