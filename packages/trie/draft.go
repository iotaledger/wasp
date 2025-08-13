package trie

import (
	"bytes"
	"fmt"
)

const KeyMaxLength = 256

// Draft is an updatable trie implemented on top of the unpackedKey/value
// store. It keeps all mutations in-memory until Commit is called.
type Draft struct {
	base        *Reader
	mutatedRoot *draftNode
}

type CommitStats struct {
	CreatedNodes  uint
	CreatedValues uint
}

func NewDraft(store KVReader, root Hash) (*Draft, error) {
	trieReader := NewReader(store, root)
	rootNodeData, ok := trieReader.nodeStore.FetchNodeData(root)
	if !ok {
		return nil, fmt.Errorf("trie root not found: %s", root)
	}
	return &Draft{
		base:        trieReader,
		mutatedRoot: newDraftNode(rootNodeData, nil),
	}, nil
}

// Update updates TrieDraft with the unpackedKey/value. Reorganizes and re-calculates trie, keeps cache consistent
func (tr *Draft) Update(key []byte, value []byte) {
	assertf(len(key) > 0, "len(key) must be > 0")
	unpackedTriePath := unpackBytes(key)
	if len(value) == 0 {
		tr.delete(unpackedTriePath)
	} else {
		tr.update(unpackedTriePath, value)
	}
}

// Delete deletes Key/value from the TrieDraft
func (tr *Draft) Delete(key []byte) {
	if len(key) == 0 {
		// we do not want to delete root
		return
	}
	tr.delete(unpackBytes(key))
}

// DeletePrefix deletes all kv pairs with the prefix. It is a very fast operation, it modifies only one node
// and all children (any number) disappears from the next root
func (tr *Draft) DeletePrefix(pathPrefix []byte) bool {
	if len(pathPrefix) == 0 {
		// we do not want to delete root, or do we?
		return false
	}
	unpackedPrefix := unpackBytes(pathPrefix)
	return tr.deletePrefix(unpackedPrefix)
}

func (tr *Draft) update(triePath []byte, value []byte) {
	assertf(len(value) > 0, "len(value) > 0")
	assertf(len(triePath) < KeyMaxLength, "len(key) = %d, must under KeyMaxLength %x", len(triePath))

	nodes := make([]*draftNode, 0)
	var ends pathEndingCode
	tr.traverseMutatedPath(triePath, func(n *draftNode, ending pathEndingCode) {
		nodes = append(nodes, n)
		ends = ending
	})
	assertf(len(nodes) > 0, "len(nodes) > 0")
	for i := len(nodes) - 2; i >= 0; i-- {
		nodes[i].setModifiedChild(nodes[i+1])
	}
	lastNode := nodes[len(nodes)-1]
	switch ends {
	case endingTerminal:
		// reached the end just for the terminal
		lastNode.setValue(value)

	case endingExtend:
		// extend the current node with the new terminal node
		keyPlusPathExtension := concat(lastNode.triePath, lastNode.pathExtension)
		assertf(len(keyPlusPathExtension) < len(triePath), "len(keyPlusPathExtension) < len(triePath)")
		childTriePath := triePath[:len(keyPlusPathExtension)+1]
		childIndex := childTriePath[len(childTriePath)-1]
		assertf(lastNode.getChild(childIndex, tr.base.nodeStore) == nil, "lastNode.getChild(childIndex, tr.nodeStore)==nil")
		child := tr.newTerminalNode(childTriePath, triePath[len(keyPlusPathExtension)+1:], value)
		lastNode.setModifiedChild(child)

	case endingSplit:
		// split the last node
		var prevNode *draftNode
		if len(nodes) >= 2 {
			prevNode = nodes[len(nodes)-2]
		}
		trieKey := lastNode.triePath
		assertf(len(trieKey) <= len(triePath), "len(trieKey) <= len(triePath)")
		remainingTriePath := triePath[len(trieKey):]

		prefix, pathExtensionTail, triePathTail := commonPrefix(lastNode.pathExtension, remainingTriePath)

		childIndexContinue := pathExtensionTail[0]
		pathExtensionContinue := pathExtensionTail[1:]
		trieKeyToContinue := concat(trieKey, prefix, []byte{childIndexContinue})

		prevNode.removeChild(lastNode)
		lastNode.setPathExtension(pathExtensionContinue)
		lastNode.setTriePath(trieKeyToContinue)

		forkingNode := newDraftNode(nil, trieKey) // will be at path of the old node
		forkingNode.setPathExtension(prefix)
		forkingNode.setModifiedChild(lastNode)
		prevNode.setModifiedChild(forkingNode)

		if len(triePathTail) == 0 {
			forkingNode.setValue(value)
		} else {
			childIndexToBranch := triePathTail[0]
			branchPathExtension := triePathTail[1:]
			trieKeyToContinue = concat(trieKey, prefix, []byte{childIndexToBranch})

			newNodeWithTerminal := tr.newTerminalNode(trieKeyToContinue, branchPathExtension, value)
			forkingNode.setModifiedChild(newNodeWithTerminal)
		}

	default:
		assertf(false, "inconsistency: wrong value")
	}
}

func (tr *Draft) newTerminalNode(triePath, pathExtension, value []byte) *draftNode {
	ret := newDraftNode(nil, triePath)
	ret.setPathExtension(pathExtension)
	ret.setValue(value)
	return ret
}

func (tr *Draft) delete(triePath []byte) {
	nodes := make([]*draftNode, 0)
	var ends pathEndingCode
	tr.traverseMutatedPath(triePath, func(n *draftNode, ending pathEndingCode) {
		nodes = append(nodes, n)
		ends = ending
	})
	assertf(len(nodes) > 0, "len(nodes) > 0")
	if ends != endingTerminal {
		// the key is not present in the trie, do nothing
		return
	}

	nodes[len(nodes)-1].setValue(nil)

	for i := len(nodes) - 1; i >= 1; i-- {
		idxAsChild := nodes[i].indexAsChild()
		n := tr.mergeNodeIfNeeded(nodes[i])
		if n != nil {
			nodes[i-1].removeChild(nodes[i])
			nodes[i-1].setModifiedChild(n)
		} else {
			nodes[i-1].removeChild(nil, idxAsChild)
		}
	}
	assertf(nodes[0] != nil, "please do not delete root")
}

// deletePrefix deletes all k/v pairs from the trie with the specified prefix
// It does nothing if prefix is nil, i.e. you can't delete the root
// return if deleted something
func (tr *Draft) deletePrefix(pathPrefix []byte) bool {
	nodes := make([]*draftNode, 0)

	prefixExists := false
	tr.traverseMutatedPath(pathPrefix, func(n *draftNode, ending pathEndingCode) {
		nodes = append(nodes, n)
		if bytes.HasPrefix(concat(n.triePath, n.nodeData.PathExtension), pathPrefix) {
			prefixExists = true
		}
	})
	if !prefixExists {
		return false
	}
	assertf(len(nodes) > 1, "len(nodes) > 0")
	// remove the last node and propagate

	// remove terminal and the children from the current node
	lastNode := nodes[len(nodes)-1]
	lastNode.setValue(nil)
	for i := range NumChildren {
		if _, isModified := lastNode.uncommittedChildren[byte(i)]; isModified {
			lastNode.uncommittedChildren[byte(i)] = nil
			continue
		}
		if c := lastNode.nodeData.Children[i]; c != nil {
			lastNode.uncommittedChildren[byte(i)] = nil
		}
	}
	for i := len(nodes) - 1; i >= 1; i-- {
		idxAsChild := nodes[i].indexAsChild()
		n := tr.mergeNodeIfNeeded(nodes[i])
		if n != nil {
			nodes[i-1].removeChild(nodes[i])
			nodes[i-1].setModifiedChild(n)
		} else {
			nodes[i-1].removeChild(nil, idxAsChild)
		}
	}
	return true
}

func (tr *Draft) mergeNodeIfNeeded(node *draftNode) *draftNode {
	toRemove, theOnlyChildToMergeWith := node.hasToBeRemoved(tr.base.nodeStore)
	if !toRemove {
		return node
	}
	if theOnlyChildToMergeWith == nil {
		// just remove
		return nil
	}
	// merge with child
	newPathExtension := concat(node.pathExtension, []byte{theOnlyChildToMergeWith.indexAsChild()}, theOnlyChildToMergeWith.pathExtension)
	theOnlyChildToMergeWith.setPathExtension(newPathExtension)
	theOnlyChildToMergeWith.setTriePath(node.triePath)
	return theOnlyChildToMergeWith
}

// Commit calculates a new mutatedRoot commitment value from the cache, commits all mutations
// and writes it into the store.
// The returned CommitStats are only valid if refcounts are enabled.
func (tr *Draft) Commit(store KVStore) (newTrieRoot Hash, refcountsEnabled bool, stats *CommitStats) {
	triePartition := makeWriterPartition(store, partitionTrieNodes)
	valuePartition := makeWriterPartition(store, partitionValues)

	commitNode(tr.mutatedRoot, triePartition, valuePartition)
	refcountsEnabled, refcounts := NewRefcounts(store)
	if refcountsEnabled {
		commitStats := refcounts.inc(tr.mutatedRoot)
		stats = &commitStats
	}

	// set uncommitted children in the root to empty -> the GC will collect the whole tree of buffered nodes
	tr.mutatedRoot.uncommittedChildren = make(map[byte]*draftNode)

	newTrieRoot = tr.mutatedRoot.nodeData.Commitment
	tr.mutatedRoot = nil // prevent future mutations using this instance of TrieDraft
	return newTrieRoot, refcountsEnabled, stats
}

// commitNode re-calculates the node commitment and, recursively, its children commitments
func commitNode(root *draftNode, triePartition, valuePartition KVWriter) {
	// traverse post-order so that we compute the commitments bottom-up
	root.traversePostOrder(func(node *draftNode) {
		childUpdates := make(map[byte]*Hash)
		for idx, child := range node.uncommittedChildren {
			if child == nil {
				childUpdates[idx] = nil
			} else {
				hashCopy := child.nodeData.Commitment
				childUpdates[idx] = &hashCopy
			}
		}
		node.nodeData.update(childUpdates, node.terminal, node.pathExtension)
		node.mustPersist(triePartition)
		if len(node.value) > 0 {
			valuePartition.Set(node.terminal.Bytes(), node.value)
		}
	})
}
