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
	rootNodeData, err := tr.TrieReader.setRoot(h)
	if err != nil {
		return err
	}
	tr.mutatedRoot = newBufferedNode(rootNodeData, nil) // the previous mutated tree will be GC-ed
	return nil
}

// Commit calculates a new mutatedRoot commitment value from the cache, commits all mutations
// and writes it into the store.
// The nodes and values are written into separate partitions
// The buffered nodes are garbage collected, except the mutated ones
// By default, it sets new root in the end and clears the trie reader cache. To override, use notSetNewRoot = true
func (tr *TrieUpdatable) Commit(store KVStore) (Hash, CommitStats) {
	triePartition := makeWriterPartition(store, partitionTrieNodes)
	valuePartition := makeWriterPartition(store, partitionValues)
	refcounts := newRefcounts(store)

	var stats CommitStats
	tr.mutatedRoot.commitNode(triePartition, valuePartition, refcounts, &stats)
	// set uncommitted children in the root to empty -> the GC will collect the whole tree of buffered nodes
	tr.mutatedRoot.uncommittedChildren = make(map[byte]*bufferedNode)

	newTrieRoot := tr.mutatedRoot.nodeData.Commitment
	err := tr.SetRoot(newTrieRoot) // always clear cache because NodeData-s are mutated and not valid anymore
	assertNoError(err)

	return newTrieRoot, stats
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
		fmt.Printf("%s %v %s\n", strings.Repeat(" ", len(nodeKey)), nodeKey, n)
		return IterateContinue
	})
}

func (tr *TrieReader) CopyToStore(snapshot KVStore) {
	triePartition := makeWriterPartition(snapshot, partitionTrieNodes)
	valuePartition := makeWriterPartition(snapshot, partitionValues)
	refcounts := newRefcounts(snapshot)
	tr.IterateNodes(func(_ []byte, n *NodeData, depth int) IterateNodesAction {
		nodeKey := n.Commitment.Bytes()
		triePartition.Set(nodeKey, tr.nodeStore.trieStore.Get(nodeKey))
		if n.Terminal != nil && !n.Terminal.IsValue {
			n.Terminal.ExtractValue()
			valueKey := n.Terminal.Bytes()
			valuePartition.Set(valueKey, tr.nodeStore.valueStore.Get(valueKey))
		}
		refcounts.incNodeAndValue(n)
		return IterateContinue
	})
}
