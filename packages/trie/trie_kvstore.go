package trie

import (
	"bytes"
)

// Get reads the value associated with the given key in the trie
func (tr *TrieRFromRoot) Get(key []byte) []byte {
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
	return tr.R.fetchValueOfTerminal(terminal)
}

// Has checks the existence of the key in the trie
func (tr *TrieRFromRoot) Has(key []byte) bool {
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
func (tr *TrieRFromRoot) Iterate(prefix []byte, f func(k []byte, v []byte) bool) {
	tr.iteratePrefix(f, prefix, true)
}

// IterateKeys iterates all the keys in the trie
func (tr *TrieRFromRoot) IterateKeys(prefix []byte, f func(k []byte) bool) {
	tr.iteratePrefix(func(k []byte, v []byte) bool { return f(k) }, prefix, false)
}

// iteratePrefix iterates the key/value with keys with prefix.
// The order of the iteration will be deterministic
func (tr *TrieRFromRoot) iteratePrefix(f func(k []byte, v []byte) bool, prefix []byte, extractValue bool) {
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
		tr.R.iterate(*root, triePath, f, extractValue)
	}
}

func (tr *TrieR) iterate(root Hash, triePath []byte, fun func(k []byte, v []byte) bool, extractValue bool) {
	rootNode, found := tr.fetchNodeData(root)
	assertf(found, "root node not found: %s", root)

	tr.iterateNodes(0, rootNode, triePath, func(nodeKey []byte, n *NodeData, depth int) IterateNodesAction {
		if n.Terminal != nil {
			key, err := packUnpackedBytes(concat(nodeKey, n.PathExtension))
			assertNoError(err)
			var value []byte
			if extractValue {
				value = tr.fetchValueOfTerminal(n.Terminal)
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
func (tr *TrieRFromRoot) IterateNodes(fun IterateNodesCallback) {
	root, found := tr.R.fetchNodeData(tr.Root)
	assertf(found, "root node not found: %s", tr.Root)
	tr.R.iterateNodes(0, root, nil, fun)
}

func (tr *TrieR) iterateNodes(depth int, n *NodeData, path []byte, fun IterateNodesCallback) bool {
	action := fun(path, n, depth)
	if action == IterateContinue {
		childrenNodes := tr.fetchChildrenNodeData(n)
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
func (tr *TrieRFromRoot) IterateNodesWithRefcounts(fun IterateNodesWithRefcountsCallback) {
	root, found := tr.R.fetchNodeData(tr.Root)
	assertf(found, "root node not found: %s", tr.Root)

	nodeRefcount := tr.R.GetNodeRefcount(root.Commitment)
	valueRefcount := uint32(0)
	if root.CommitsToExternalValue() {
		valueRefcount = tr.R.GetValueRefcount(root.Terminal)
	}
	tr.R.iterateNodesWithRefcounts(0, root, nil, nodeRefcount, valueRefcount, fun)
}

func (tr *TrieR) iterateNodesWithRefcounts(depth int, n *NodeData, path []byte, nodeRefcount, valueRefcount uint32, fun IterateNodesWithRefcountsCallback) bool {
	action := fun(path, n, depth, nodeRefcount, valueRefcount)
	if action == IterateContinue {
		childrenNodes := tr.fetchChildrenNodeDataWithRefcounts(n)
		for childIndex := range NumChildren {
			if childrenNodes[childIndex].node == nil {
				continue
			}
			if !tr.iterateNodesWithRefcounts(
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
