package trie

import (
	"fmt"
)

// TrieUpdatable is an updatable trie implemented on top of the unpackedKey/value store. It is virtualized and optimized by caching of the
// trie update operation and keeping consistent trie in the cache
type TrieUpdatable struct {
	*TrieReader
	mutatedRoot *bufferedNode
}

// TrieReader direct read-only access to trie
type TrieReader struct {
	nodeStore      *nodeStore
	persistentRoot VCommitment
}

func NewTrieUpdatable(store KVReader, root VCommitment, clearCacheAtSize ...int) (*TrieUpdatable, error) {
	trieReader, err := NewTrieReader(store, root, clearCacheAtSize...)
	if err != nil {
		return nil, err
	}
	ret := &TrieUpdatable{
		TrieReader: trieReader,
	}
	if err = ret.setRoot(root); err != nil {
		return nil, err
	}
	return ret, nil
}

func NewTrieReader(store KVReader, root VCommitment, clearCacheAtSize ...int) (*TrieReader, error) {
	ret := &TrieReader{
		nodeStore: openNodeStore(store, clearCacheAtSize...),
	}
	if _, err := ret.setRoot(root); err != nil {
		return nil, err
	}
	return ret, nil
}

func (tr *TrieReader) Root() VCommitment {
	return tr.persistentRoot
}

func (tr *TrieReader) ClearCache() {
	tr.nodeStore.clearCache()
}

// SetRoot fetches and sets new root. It clears cache before fetching the new root
func (tr *TrieReader) setRoot(c VCommitment) (*nodeData, error) {
	tr.ClearCache()
	rootNodeData, ok := tr.nodeStore.FetchNodeData(c)
	if !ok {
		return nil, fmt.Errorf("root commitment '%s' does not exist", c)
	}
	tr.persistentRoot = c.Clone()
	return rootNodeData, nil
}

// setRoot overloaded for updatable trie
func (tr *TrieUpdatable) setRoot(c VCommitment) error {
	rootNodeData, err := tr.TrieReader.setRoot(c)
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
func (tr *TrieUpdatable) Commit(store KVWriter) VCommitment {
	triePartition := makeWriterPartition(store, partitionTrieNodes)
	valuePartition := makeWriterPartition(store, partitionValues)

	tr.mutatedRoot.commitNode(triePartition, valuePartition)
	// set uncommitted children in the root to empty -> the GC will collect the whole tree of buffered nodes
	tr.mutatedRoot.uncommittedChildren = make(map[byte]*bufferedNode)

	ret := tr.mutatedRoot.nodeData.commitment.Clone()
	err := tr.setRoot(ret) // always clear cache because NodeData-s are mutated and not valid anymore
	assertNoError(err)
	return ret
}

func (tr *TrieUpdatable) newTerminalNode(triePath, pathExtension, value []byte) *bufferedNode {
	ret := newBufferedNode(nil, triePath)
	ret.setPathExtension(pathExtension)
	ret.setValue(value)
	return ret
}
