package trie

import (
	"fmt"
	"strings"
)

// TrieUpdatable is an updatable trie implemented on top of the unpackedKey/value store. It is virtualized and optimized by caching of the
// trie update operation and keeping consistent trie in the cache
type TrieUpdatable struct {
	*TrieReader
	mutatedRoot *bufferedNode
}

// TrieReader direct read-only access to trie
type TrieReader struct {
	nodeStore *nodeStore
	root      Hash
}

type CommitStats struct {
	CreatedNodes  uint
	CreatedValues uint
}

func NewTrieUpdatable(store KVReader, root Hash) (*TrieUpdatable, error) {
	trieReader, err := NewTrieReader(store, root)
	if err != nil {
		return nil, err
	}
	ret := &TrieUpdatable{
		TrieReader: trieReader,
	}
	if err := ret.SetRoot(root); err != nil {
		return nil, err
	}
	return ret, nil
}

func NewTrieReader(store KVReader, root Hash) (*TrieReader, error) {
	ret := &TrieReader{
		nodeStore: openNodeStore(store),
	}
	if _, err := ret.setRoot(root); err != nil {
		return nil, err
	}
	return ret, nil
}

func (tr *TrieReader) Root() Hash {
	return tr.root
}

// SetRoot fetches and sets new root. It clears cache before fetching the new root
func (tr *TrieReader) setRoot(h Hash) (*NodeData, error) {
	rootNodeData, ok := tr.nodeStore.FetchNodeData(h)
	if !ok {
		return nil, fmt.Errorf("root commitment '%s' does not exist", &h)
	}
	tr.root = h
	return rootNodeData, nil
}

// SetRoot overloaded for updatable trie
func (tr *TrieUpdatable) SetRoot(h Hash) error {
	rootNodeData, err := tr.setRoot(h)
	if err != nil {
		return err
	}
	tr.mutatedRoot = newBufferedNode(rootNodeData, nil) // the previous mutated tree will be GC-ed
	return nil
}

// Commit calculates a new mutatedRoot commitment value from the cache, commits all mutations
// and writes it into the store.
// The returned CommitStats are only valid if refcounts are enabled.
func (tr *TrieUpdatable) Commit(store KVStore) (newTrieRoot Hash, refcountsEnabled bool, stats *CommitStats) {
	triePartition := makeWriterPartition(store, partitionTrieNodes)
	valuePartition := makeWriterPartition(store, partitionValues)

	commitNode(tr.mutatedRoot, triePartition, valuePartition)
	refcountsEnabled, refcounts := NewRefcounts(store)
	if refcountsEnabled {
		commitStats := refcounts.inc(tr.mutatedRoot)
		stats = &commitStats
	}

	// set uncommitted children in the root to empty -> the GC will collect the whole tree of buffered nodes
	tr.mutatedRoot.uncommittedChildren = make(map[byte]*bufferedNode)

	newTrieRoot = tr.mutatedRoot.nodeData.Commitment
	err := tr.SetRoot(newTrieRoot) // always clear cache because NodeData-s are mutated and not valid anymore
	assertNoError(err)

	return newTrieRoot, refcountsEnabled, stats
}

// commitNode re-calculates the node commitment and, recursively, its children commitments
func commitNode(root *bufferedNode, triePartition, valuePartition KVWriter) {
	// traverse post-order so that we compute the commitments bottom-up
	root.traversePostOrder(func(node *bufferedNode) {
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

func (tr *TrieUpdatable) newTerminalNode(triePath, pathExtension, value []byte) *bufferedNode {
	ret := newBufferedNode(nil, triePath)
	ret.setPathExtension(pathExtension)
	ret.setValue(value)
	return ret
}

// DebugDump prints the structure of the tree to stdout, for debugging purposes.
func (tr *TrieReader) DebugDump() {
	tr.IterateNodes(func(nodeKey []byte, n *NodeData, depth int) IterateNodesAction {
		key := "[]"
		if len(nodeKey) > 0 {
			key = fmt.Sprintf("[%d]", nodeKey[len(nodeKey)-1])
		}
		indent := strings.Repeat(" ", depth*4)
		fmt.Printf("%s %v %s\n", indent, key, n)
		if n.Terminal != nil && !n.Terminal.IsValue {
			fmt.Printf(
				"%s     [v: %x -> %q]\n",
				indent,
				n.Terminal.Data,
				ellipsis(tr.nodeStore.valueStore.Get(n.Terminal.Bytes()), 20),
			)
		}
		return IterateContinue
	})
}

func ellipsis(b []byte, maxLen int) string {
	if len(b) <= maxLen {
		return string(b)
	}
	if maxLen < 3 {
		maxLen = 3
	}
	return string(b[0:maxLen-3]) + "..."
}

// DebugDump prints the structure of the whole DB to stdout, for debugging purposes.
func DebugDump(store KVStore, roots []Hash) {
	fmt.Printf("[trie store]\n")
	for _, root := range roots {
		tr, err := NewTrieReader(store, root)
		assertNoError(err)
		tr.DebugDump()
	}
	enabled, refcounts := NewRefcounts(store)
	if enabled {
		refcounts.DebugDump()
	} else {
		fmt.Printf("[node refcounts disabled]\n")
	}
}
