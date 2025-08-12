package trie

import (
	"encoding/hex"

	"github.com/iotaledger/wasp/v2/packages/kv/codec"
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
	partitionRefcountsEnabled
)

// InitRoot initializes a new empty trie
func InitRoot(store KVStore, refcountsEnabled bool) (Hash, error) {
	err := UpdateRefcountsFlag(store, refcountsEnabled)
	if err != nil {
		return Hash{}, err
	}

	rootNodeData := newNodeData()
	n := newBufferedNode(rootNodeData, nil)

	trieStore := makeWriterPartition(store, partitionTrieNodes)
	valueStore := makeWriterPartition(store, partitionValues)
	commitNode(n, trieStore, valueStore)
	initRefcounts(store, n)
	return n.nodeData.Commitment, nil
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

// fetchChildrenNodeData fetches the children nodes of a given node in a single call to MultiGet.
func (ns *nodeStore) fetchChildrenNodeData(n *NodeData) [NumChildren]*NodeData {
	// fetch using a single call to MultiGet
	dbKeys := make([][]byte, 0, NumChildren)
	for _, hash := range n.Children {
		if hash == nil {
			continue
		}
		dbKeys = append(dbKeys, hash.Bytes())
	}

	children := [NumChildren]*NodeData{}
	if len(dbKeys) == 0 {
		// nothing to fetch
		return children
	}

	nodeBins := ns.trieStore.MultiGet(dbKeys)

	// decode results
	for i, hash := range n.Children {
		if hash == nil {
			continue
		}
		nodeBin := nodeBins[0]
		nodeBins = nodeBins[1:]

		assertf(nodeBin != nil, "NodeStore::FetchChildrenNodeData: nodeBin is nil for child index %d, commitment: %s",
			i, n.Commitment.String())
		nodeData, err := nodeDataFromBytes(nodeBin)
		assertf(err == nil, "NodeStore::FetchChildrenNodeData err: '%v' nodeBin: '%s', commitment: %x",
			err, hex.EncodeToString(nodeBin), *hash)
		nodeData.Commitment = *hash
		children[i] = nodeData
	}
	return children
}

type NodeDataWithRefcounts struct {
	node          *NodeData
	nodeRefcount  uint32
	valueRefcount uint32
}

// fetchChildrenNodeDataWithRefcounts fetches the children nodes and their refcounts in two MultiGet calls.
func (ns *nodeStore) fetchChildrenNodeDataWithRefcounts(refcounts *Refcounts, n *NodeData) [NumChildren]NodeDataWithRefcounts {
	// fetch nodes with MultiGet
	nodeDatas := ns.fetchChildrenNodeData(n)
	// now fetch their refcounts with another MultiGet
	dbKeys := make([][]byte, 0, NumChildren*2)
	for _, child := range nodeDatas {
		if child == nil {
			continue
		}
		dbKeys = append(dbKeys, nodeRefcountKey(child.Commitment[:]))
		if child.Terminal != nil && !child.Terminal.IsValue {
			dbKeys = append(dbKeys, valueRefcountKey(child.Terminal.Data))
		}
	}
	var ret [NumChildren]NodeDataWithRefcounts
	if len(dbKeys) == 0 {
		// nothing to fetch
		return ret
	}
	refCountBytes := refcounts.store.MultiGet(dbKeys)
	for i, child := range nodeDatas {
		if child == nil {
			continue
		}
		nodeRefcount := codec.MustDecode[uint32](refCountBytes[0], 0)
		refCountBytes = refCountBytes[1:]
		valueRefcount := uint32(0)
		if child.Terminal != nil && !child.Terminal.IsValue {
			valueRefcount = codec.MustDecode[uint32](refCountBytes[0], 0)
			refCountBytes = refCountBytes[1:]
		}
		ret[i] = NodeDataWithRefcounts{
			node:          child,
			nodeRefcount:  nodeRefcount,
			valueRefcount: valueRefcount,
		}
	}
	return ret
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
