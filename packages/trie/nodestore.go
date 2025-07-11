package trie

import (
	"encoding/hex"

	"github.com/samber/lo"
)

// nodeStore immutable node store
type nodeStore struct {
	trieStore  KVReader
	valueStore KVReader
}

const (
	partitionTrieNodes = byte(iota)
	partitionValues
	partitionRefcountNodes
	partitionRefcountValues
)

// MustInitRoot initializes a new empty trie
func MustInitRoot(store KVStore) Hash {
	rootNodeData := newNodeData()
	n := newBufferedNode(rootNodeData, nil)

	trieStore := makeWriterPartition(store, partitionTrieNodes)
	valueStore := makeWriterPartition(store, partitionValues)
	refcounts := newRefcounts(store)
	var stats CommitStats
	n.commitNode(trieStore, valueStore, refcounts, &stats)

	return n.nodeData.Commitment
}

func openNodeStore(store KVReader) *nodeStore {
	store = makeCachedKVReader(store)
	return &nodeStore{
		trieStore:  makeReaderPartition(store, partitionTrieNodes),
		valueStore: makeReaderPartition(store, partitionValues),
	}
}

func (ns *nodeStore) FetchNodeData(nodeCommitment Hash) (*NodeData, bool) {
	dbKey := nodeCommitment.Bytes()
	nodeBin := ns.trieStore.Get(dbKey)
	if len(nodeBin) == 0 {
		return nil, false
	}
	ret, err := nodeDataFromBytes(nodeBin)
	assertf(err == nil, "NodeStore::FetchNodeData err: '%v' nodeBin: '%s', commitment: %s",
		err, hex.EncodeToString(nodeBin), nodeCommitment)
	ret.Commitment = nodeCommitment
	return ret, true
}

func (ns *nodeStore) FetchChildrenNodeData(n *NodeData) [NumChildren]*NodeData {
	// fetch using a single call to MultiGet
	var dbKeys [NumChildren][]byte
	var positions [NumChildren]int
	numChildren := 0
	for i := range NumChildren {
		hash := n.Children[i]
		if hash == nil {
			continue
		}
		dbKeys[numChildren] = hash.Bytes()
		positions[i] = numChildren
		numChildren++
	}

	children := [NumChildren]*NodeData{}
	if numChildren == 0 {
		// nothing to fetch
		return children
	}

	nodeBins := ns.trieStore.MultiGet(dbKeys[:numChildren])

	// decode results
	for i := range NumChildren {
		if n.Children[i] == nil {
			continue
		}
		p := positions[i]
		nodeBin := nodeBins[p]
		assertf(nodeBin != nil, "NodeStore::FetchChildrenNodeData: nodeBin is nil for child index %d, commitment: %s",
			i, n.Commitment.String())
		nodeData, err := nodeDataFromBytes(nodeBin)
		assertf(err == nil, "NodeStore::FetchChildrenNodeData err: '%v' nodeBin: '%s', commitment: %x",
			err, hex.EncodeToString(nodeBin), dbKeys[p])
		nodeData.Commitment = lo.Must(HashFromBytes(dbKeys[p]))
		children[i] = nodeData
	}
	return children
}

func (ns *nodeStore) MustFetchNodeData(nodeCommitment Hash) *NodeData {
	ret, ok := ns.FetchNodeData(nodeCommitment)
	assertf(ok, "NodeStore::MustFetchNodeData: cannot find node data: commitment: '%s'", nodeCommitment.String())
	return ret
}

func (ns *nodeStore) FetchChild(n *NodeData, childIdx byte, trieKey []byte) (*NodeData, []byte) {
	c := n.Children[childIdx]
	if c == nil {
		return nil, nil
	}
	childTriePath := concat(trieKey, n.PathExtension, []byte{childIdx})

	ret, ok := ns.FetchNodeData(*c)
	assertf(ok, "immutable::FetchChild: failed to fetch node. trieKey: '%s', childIndex: %d",
		hex.EncodeToString(trieKey), childIdx)
	return ret, childTriePath
}
