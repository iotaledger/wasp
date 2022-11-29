package trie

import (
	"encoding/hex"
)

// nodeStore immutable node store
type nodeStore struct {
	trieStore        KVReader
	valueStore       KVReader
	cache            map[string]*nodeData
	clearCacheAtSize int
}

const defaultClearCacheEveryGets = 1000

const (
	partitionTrieNodes = byte(iota)
	partitionValues
)

// MustInitRoot initializes a new empty trie
func MustInitRoot(store KVWriter) Hash {
	rootNodeData := newNodeData()
	n := newBufferedNode(rootNodeData, nil)

	trieStore := makeWriterPartition(store, partitionTrieNodes)
	valueStore := makeWriterPartition(store, partitionValues)
	n.commitNode(trieStore, valueStore)

	return n.nodeData.commitment
}

func openNodeStore(store KVReader, clearCacheAtSize ...int) *nodeStore {
	ret := &nodeStore{
		trieStore:        makeReaderPartition(store, partitionTrieNodes),
		valueStore:       makeReaderPartition(store, partitionValues),
		cache:            make(map[string]*nodeData),
		clearCacheAtSize: defaultClearCacheEveryGets,
	}
	if len(clearCacheAtSize) > 0 {
		ret.clearCacheAtSize = clearCacheAtSize[0]
	}
	return ret
}

func (ns *nodeStore) FetchNodeData(nodeCommitment Hash) (*nodeData, bool) {
	dbKey := nodeCommitment.Bytes()
	if ns.clearCacheAtSize > 0 {
		// if caching is used at all
		if ret, inCache := ns.cache[string(dbKey)]; inCache {
			return ret, true
		}
		if len(ns.cache) > ns.clearCacheAtSize {
			// GC the whole cache when cache reaches specified size
			// TODO: improve
			ns.cache = make(map[string]*nodeData)
		}
	}
	nodeBin := ns.trieStore.Get(dbKey)
	if len(nodeBin) == 0 {
		return nil, false
	}
	ret, err := nodeDataFromBytes(nodeBin)
	assert(err == nil, "NodeStore::FetchNodeData err: '%v' nodeBin: '%s', commitment: %s",
		err, hex.EncodeToString(nodeBin), nodeCommitment)
	ret.commitment = nodeCommitment
	return ret, true
}

func (ns *nodeStore) MustFetchNodeData(nodeCommitment Hash) *nodeData {
	ret, ok := ns.FetchNodeData(nodeCommitment)
	assert(ok, "NodeStore::MustFetchNodeData: cannot find node data: commitment: '%s'", nodeCommitment.String())
	return ret
}

func (ns *nodeStore) FetchChild(n *nodeData, childIdx byte, trieKey []byte) (*nodeData, []byte) {
	c := n.children[childIdx]
	if c == nil {
		return nil, nil
	}
	childTriePath := concat(trieKey, n.pathExtension, []byte{childIdx})

	ret, ok := ns.FetchNodeData(*c)
	assert(ok, "immutable::FetchChild: failed to fetch node. trieKey: '%s', childIndex: %d",
		hex.EncodeToString(trieKey), childIdx)
	return ret, childTriePath
}

func (ns *nodeStore) clearCache() {
	ns.cache = make(map[string]*nodeData)
}
