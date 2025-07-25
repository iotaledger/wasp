package trie

import (
	"encoding/hex"
)

// bufferedNode is a modified node
type bufferedNode struct {
	// persistent
	nodeData            *NodeData
	value               []byte // will be persisted in value store if not nil
	terminal            *Tcommitment
	pathExtension       []byte
	uncommittedChildren map[byte]*bufferedNode // children which has been modified
	triePath            []byte
}

func newBufferedNode(n *NodeData, triePath []byte) *bufferedNode {
	if n == nil {
		n = newNodeData()
	}
	ret := &bufferedNode{
		nodeData:            n,
		terminal:            n.Terminal,
		pathExtension:       n.PathExtension,
		uncommittedChildren: make(map[byte]*bufferedNode),
		triePath:            triePath,
	}
	return ret
}

func (n *bufferedNode) mustPersist(partition KVWriter) {
	dbKey := n.nodeData.Commitment.Bytes()
	partition.Set(dbKey, n.nodeData.Bytes())
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
	n.terminal = CommitToData(value)
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
	childCommitment := n.nodeData.Children[childIndex]
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

	for i := range byte(NumChildren) {
		child := n.getChild(i, nodeStore)
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
func (n *bufferedNode) traversePreOrder(f func(*bufferedNode) IterateNodesAction) bool {
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
func (n *bufferedNode) traversePostOrder(f func(*bufferedNode)) {
	for i := range byte(NumChildren) {
		if child := n.uncommittedChildren[i]; child != nil {
			child.traversePostOrder(f)
		}
	}
	f(n)
}
