package trie

import (
	"bytes"
	"encoding/hex"
)

// Get reads the trie with the key
func (tr *Reader) Get(key []byte) []byte {
	unpackedTriePath := unpackBytes(key)
	var terminal *Tcommitment
	tr.traversePath(unpackedTriePath, func(n *NodeData, _ []byte, ending pathEndingCode) {
		if ending == endingTerminal && n.Terminal != nil {
			terminal = n.Terminal
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
func (tr *Reader) Has(key []byte) bool {
	unpackedTriePath := unpackBytes(key)
	found := false
	tr.traversePath(unpackedTriePath, func(n *NodeData, p []byte, ending pathEndingCode) {
		if ending == endingTerminal && n.Terminal != nil {
			found = true
		}
	})
	return found
}

// Iterate iterates all the key/value pairs in the trie
func (tr *Reader) Iterate(prefix []byte, f func(k []byte, v []byte) bool) {
	tr.iteratePrefix(f, prefix, true)
}

// IterateKeys iterates all the keys in the trie
func (tr *Reader) IterateKeys(prefix []byte, f func(k []byte) bool) {
	tr.iteratePrefix(func(k []byte, v []byte) bool { return f(k) }, prefix, false)
}

// iteratePrefix iterates the key/value with keys with prefix.
// The order of the iteration will be deterministic
func (tr *Reader) iteratePrefix(f func(k []byte, v []byte) bool, prefix []byte, extractValue bool) {
	var root *Hash
	var triePath []byte
	unpackedPrefix := unpackBytes(prefix)
	tr.traversePath(unpackedPrefix, func(n *NodeData, trieKey []byte, ending pathEndingCode) {
		if bytes.HasPrefix(concat(trieKey, n.PathExtension), unpackedPrefix) {
			root = &n.Commitment
			triePath = trieKey
		}
	})
	if root != nil {
		tr.iterate(*root, triePath, f, extractValue)
	}
}

func (tr *Reader) iterate(root Hash, triePath []byte, fun func(k []byte, v []byte) bool, extractValue bool) {
	rootNode, found := tr.nodeStore.FetchNodeData(root)
	assertf(found, "root node not found: %s", tr.root)

	tr.iterateNodes(0, rootNode, triePath, func(nodeKey []byte, n *NodeData, depth int) IterateNodesAction {
		if n.Terminal != nil {
			key, err := packUnpackedBytes(concat(nodeKey, n.PathExtension))
			assertNoError(err)
			var value []byte
			if extractValue {
				var inTheCommitment bool
				value, inTheCommitment = n.Terminal.ExtractValue()
				if !inTheCommitment {
					value = tr.nodeStore.valueStore.Get(n.Terminal.Bytes())
					assertf(len(value) > 0, "can't fetch value. triePath: '%s', data commitment: %s", hex.EncodeToString(key), n.Terminal)
				}
			}
			if !fun(key, value) {
				return IterateStop
			}
		}
		return IterateContinue
	})
}

type (
	IterateNodesAction                byte
	IterateNodesCallback              = func(nodeKey []byte, n *NodeData, depth int) IterateNodesAction
	IterateNodesWithRefcountsCallback = func(nodeKey []byte, n *NodeData, depth int, nodeRefcount, valueRefcount uint32) IterateNodesAction
)

const (
	IterateStop IterateNodesAction = iota
	IterateContinue
	IterateSkipSubtree
)

// IterateNodes iterates nodes of the trie in the lexicographical order of trie keys in "depth first" order
func (tr *Reader) IterateNodes(fun IterateNodesCallback) {
	root, found := tr.nodeStore.FetchNodeData(tr.root)
	assertf(found, "root node not found: %s", tr.root)

	tr.iterateNodes(0, root, nil, fun)
}

func (tr *Reader) iterateNodes(depth int, n *NodeData, path []byte, fun IterateNodesCallback) bool {
	action := fun(path, n, depth)
	if action == IterateContinue {
		childrenNodes := tr.nodeStore.fetchChildrenNodeData(n)
		for childIndex := range NumChildren {
			if childrenNodes[childIndex] == nil {
				continue
			}
			if !tr.iterateNodes(depth+1, childrenNodes[childIndex], concat(path, n.PathExtension, []byte{byte(childIndex)}), fun) {
				break
			}
		}
	}
	return action != IterateStop
}

// IterateNodesWithRefcounts is like IterateNodes but it also fetches each node's refcount
func (tr *Reader) IterateNodesWithRefcounts(refcounts *Refcounts, fun IterateNodesWithRefcountsCallback) {
	root, found := tr.nodeStore.FetchNodeData(tr.root)
	assertf(found, "root node not found: %s", tr.root)

	nodeRefcount := refcounts.GetNode(root.Commitment)
	valueRefcount := uint32(0)
	if root.Terminal != nil && !root.Terminal.IsValue {
		valueRefcount = refcounts.GetValue(root.Terminal.Bytes())
	}
	tr.iterateNodesWithRefcounts(refcounts, 0, root, nil, nodeRefcount, valueRefcount, fun)
}

func (tr *Reader) iterateNodesWithRefcounts(refcounts *Refcounts, depth int, n *NodeData, path []byte, nodeRefcount, valueRefcount uint32, fun IterateNodesWithRefcountsCallback) bool {
	action := fun(path, n, depth, nodeRefcount, valueRefcount)
	if action == IterateContinue {
		childrenNodes := tr.nodeStore.fetchChildrenNodeDataWithRefcounts(refcounts, n)
		for childIndex := range NumChildren {
			if childrenNodes[childIndex].node == nil {
				continue
			}
			if !tr.iterateNodesWithRefcounts(
				refcounts,
				depth+1,
				childrenNodes[childIndex].node,
				concat(path, n.PathExtension, []byte{byte(childIndex)}),
				childrenNodes[childIndex].nodeRefcount,
				childrenNodes[childIndex].valueRefcount,
				fun,
			) {
				break
			}
		}
	}
	return action != IterateStop
}
