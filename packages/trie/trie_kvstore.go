package trie

import (
	"bytes"
	"encoding/hex"
)

// Update updates TrieUpdatable with the unpackedKey/value. Reorganizes and re-calculates trie, keeps cache consistent
func (tr *TrieUpdatable) Update(key []byte, value []byte) {
	assertf(len(key) > 0, "len(key) must be > 0")
	unpackedTriePath := unpackBytes(key)
	if len(value) == 0 {
		tr.delete(unpackedTriePath)
	} else {
		tr.update(unpackedTriePath, value)
	}
}

// Delete deletes Key/value from the TrieUpdatable
func (tr *TrieUpdatable) Delete(key []byte) {
	if len(key) == 0 {
		// we do not want to delete root
		return
	}
	tr.delete(unpackBytes(key))
}

// DeletePrefix deletes all kv pairs with the prefix. It is a very fast operation, it modifies only one node
// and all children (any number) disappears from the next root
func (tr *TrieUpdatable) DeletePrefix(pathPrefix []byte) bool {
	if len(pathPrefix) == 0 {
		// we do not want to delete root, or do we?
		return false
	}
	unpackedPrefix := unpackBytes(pathPrefix)
	return tr.deletePrefix(unpackedPrefix)
}

// Get reads the trie with the key
func (tr *TrieReader) Get(key []byte) []byte {
	unpackedTriePath := unpackBytes(key)
	var terminal *tcommitment
	tr.traversePath(unpackedTriePath, func(n *nodeData, _ []byte, ending pathEndingCode) {
		if ending == endingTerminal && n.terminal != nil {
			terminal = n.terminal
		}
	})
	if terminal == nil {
		return nil
	}
	value, valueInCommitment := terminal.ExtractValue()
	if valueInCommitment {
		assertf(len(value) > 0, "value in commitment must be not nil. Unpacked key: '%s'",
			hex.EncodeToString(unpackedTriePath))
		return value
	}
	value = tr.nodeStore.valueStore.Get(terminal.Bytes())
	assertf(len(value) > 0, "value in the value store must be not nil. Unpacked key: '%s'",
		hex.EncodeToString(unpackedTriePath))
	return value
}

// Has check existence of the key in the trie
func (tr *TrieReader) Has(key []byte) bool {
	unpackedTriePath := unpackBytes(key)
	found := false
	tr.traversePath(unpackedTriePath, func(n *nodeData, p []byte, ending pathEndingCode) {
		if ending == endingTerminal && n.terminal != nil {
			found = true
		}
	})
	return found
}

// Iterate iterates all the key/value pairs in the trie
func (tr *TrieReader) Iterate(f func(k []byte, v []byte) bool) {
	tr.iteratePrefix(f, nil, true)
}

// IterateKeys iterates all the keys in the trie
func (tr *TrieReader) IterateKeys(f func(k []byte) bool) {
	tr.iteratePrefix(func(k []byte, v []byte) bool { return f(k) }, nil, false)
}

// TrieIterator implements KVIterator interface for keys in the trie with given prefix
type TrieIterator struct {
	prefix []byte
	tr     *TrieReader
}

func (ti *TrieIterator) Iterate(fun func(k []byte, v []byte) bool) {
	ti.tr.iteratePrefix(fun, ti.prefix, true)
}

func (ti *TrieIterator) IterateKeys(fun func(k []byte) bool) {
	ti.tr.iteratePrefix(func(k []byte, v []byte) bool { return fun(k) }, ti.prefix, false)
}

// Iterator returns iterator for the sub-trie
func (tr *TrieReader) Iterator(prefix []byte) KVIterator {
	return &TrieIterator{
		prefix: prefix,
		tr:     tr,
	}
}

// SnapshotData writes all key/value pairs, committed in the specific root, to a store
func (tr *TrieReader) SnapshotData(dest KVWriter) {
	tr.Iterate(func(k []byte, v []byte) bool {
		dest.Set(k, v)
		return true
	})
}

// Snapshot writes the whole trie (including values) from specific root to another store
func (tr *TrieReader) Snapshot(destStore KVWriter) {
	triePartition := makeWriterPartition(destStore, partitionTrieNodes)
	valuePartition := makeWriterPartition(destStore, partitionValues)

	tr.iterateNodes(tr.root, nil, func(nodeKey []byte, n *nodeData) bool {
		// write trie node
		var buf bytes.Buffer
		err := n.Write(&buf)
		assertNoError(err)
		triePartition.Set(n.commitment.Bytes(), buf.Bytes())

		if n.terminal == nil {
			return true
		}
		// write value if needed
		if _, valueInCommitment := n.terminal.ExtractValue(); valueInCommitment {
			return true
		}
		valueKey := n.terminal.Bytes()
		value := tr.nodeStore.valueStore.Get(valueKey)
		assertf(len(value) > 0, "can't find value for nodeKey '%s'", hex.EncodeToString(valueKey))
		valuePartition.Set(valueKey, value)
		return true
	})
}

func (tr *TrieUpdatable) update(triePath []byte, value []byte) {
	assertf(len(value) > 0, "len(value)>0")

	nodes := make([]*bufferedNode, 0)
	var ends pathEndingCode
	tr.traverseMutatedPath(triePath, func(n *bufferedNode, ending pathEndingCode) {
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
		assertf(lastNode.getChild(childIndex, tr.nodeStore) == nil, "lastNode.getChild(childIndex, tr.nodeStore)==nil")
		child := tr.newTerminalNode(childTriePath, triePath[len(keyPlusPathExtension)+1:], value)
		lastNode.setModifiedChild(child)

	case endingSplit:
		// split the last node
		var prevNode *bufferedNode
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

		forkingNode := newBufferedNode(nil, trieKey) // will be at path of the old node
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

func (tr *TrieUpdatable) delete(triePath []byte) {
	nodes := make([]*bufferedNode, 0)
	var ends pathEndingCode
	tr.traverseMutatedPath(triePath, func(n *bufferedNode, ending pathEndingCode) {
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

func (tr *TrieUpdatable) mergeNodeIfNeeded(node *bufferedNode) *bufferedNode {
	toRemove, theOnlyChildToMergeWith := node.hasToBeRemoved(tr.nodeStore)
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

// iteratePrefix iterates the key/value with keys with prefix.
// The order of the iteration will be deterministic
func (tr *TrieReader) iteratePrefix(f func(k []byte, v []byte) bool, prefix []byte, extractValue bool) {
	var root *Hash
	var triePath []byte
	unpackedPrefix := unpackBytes(prefix)
	tr.traversePath(unpackedPrefix, func(n *nodeData, trieKey []byte, ending pathEndingCode) {
		if bytes.HasPrefix(concat(trieKey, n.pathExtension), unpackedPrefix) {
			root = &n.commitment
			triePath = trieKey
		}
	})
	if root != nil {
		tr.iterate(*root, triePath, f, extractValue)
	}
}

func (tr *TrieReader) iterate(root Hash, triePath []byte, fun func(k []byte, v []byte) bool, extractValue bool) bool {
	return tr.iterateNodes(root, triePath, func(nodeKey []byte, n *nodeData) bool {
		if n.terminal != nil {
			key, err := packUnpackedBytes(concat(nodeKey, n.pathExtension))
			assertNoError(err)
			var value []byte
			if extractValue {
				var inTheCommitment bool
				value, inTheCommitment = n.terminal.ExtractValue()
				if !inTheCommitment {
					value = tr.nodeStore.valueStore.Get(n.terminal.Bytes())
					assertf(len(value) > 0, "can't fetch value. triePath: '%s', data commitment: %s", hex.EncodeToString(key), n.terminal)
				}
			}
			if !fun(key, value) {
				return false
			}
		}
		return true
	})
}

// iterateNodes iterates nodes of the trie in the lexicographical order of trie keys in "depth first" order
func (tr *TrieReader) iterateNodes(root Hash, rootKey []byte, fun func(nodeKey []byte, n *nodeData) bool) bool {
	n, found := tr.nodeStore.FetchNodeData(root)
	assertf(found, "can't fetch node. triePath: '%s', node commitment: %s", hex.EncodeToString(rootKey), root)

	if !fun(rootKey, n) {
		return false
	}
	return n.iterateChildren(func(childIndex byte, childCommitment Hash) bool {
		return tr.iterateNodes(childCommitment, concat(rootKey, n.pathExtension, []byte{childIndex}), fun)
	})
}

// deletePrefix deletes all k/v pairs from the trie with the specified prefix
// It does nothing if prefix is nil, i.e. you can't delete the root
// return if deleted something
func (tr *TrieUpdatable) deletePrefix(pathPrefix []byte) bool {
	nodes := make([]*bufferedNode, 0)

	prefixExists := false
	tr.traverseMutatedPath(pathPrefix, func(n *bufferedNode, ending pathEndingCode) {
		nodes = append(nodes, n)
		if bytes.HasPrefix(concat(n.triePath, n.nodeData.pathExtension), pathPrefix) {
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
	for i := 0; i < NumChildren; i++ {
		if _, isModified := lastNode.uncommittedChildren[byte(i)]; isModified {
			lastNode.uncommittedChildren[byte(i)] = nil
			continue
		}
		if c := lastNode.nodeData.children[i]; c != nil {
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

// utility functions for testing

func (tr *TrieReader) GetStr(key string) string {
	return string(tr.Get([]byte(key)))
}

func (tr *TrieReader) HasStr(key string) bool {
	return tr.Has([]byte(key))
}

// UpdateStr updates key/value pair in the trie
func (tr *TrieUpdatable) UpdateStr(key interface{}, value interface{}) {
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

// DeleteStr removes key from trie
func (tr *TrieUpdatable) DeleteStr(key interface{}) {
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
