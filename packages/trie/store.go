package trie

import (
	"encoding/hex"

	"github.com/iotaledger/wasp/v2/packages/kv/codec"
)

const (
	partitionTrieNodes = byte(iota)
	partitionValues
	partitionRefcountNodes
	partitionRefcountValues
	partitionRefcountsEnabled
)

func dbKeyNodeData(nodeCommitment Hash) []byte {
	return append([]byte{partitionTrieNodes}, nodeCommitment.Bytes()...)
}

func dbKeyValue(terminalBytes []byte) []byte {
	return append([]byte{partitionValues}, terminalBytes...)
}

func dbKeyNodeRefcount(nodeCommitment []byte) []byte {
	return append([]byte{partitionRefcountNodes}, nodeCommitment...)
}

func dbKeyValueRefcount(terminalData []byte) []byte {
	return append([]byte{partitionRefcountValues}, terminalData...)
}

func (terminal *Tcommitment) dbKeyValue() []byte {
	assertf(!terminal.IsValue, "dbKeyValue called on non-external value")
	return dbKeyValue(terminal.Bytes())
}

func (terminal *Tcommitment) dbKeyValueRefcount() []byte {
	assertf(!terminal.IsValue, "dbKeyValueRefcount called on non-external value")
	return dbKeyValueRefcount(terminal.Data)
}

func (n *NodeData) dbKey() []byte {
	return dbKeyNodeData(n.Commitment)
}

func (n *NodeData) dbKeyNodeRefcount() []byte {
	return dbKeyNodeRefcount(n.Commitment[:])
}

func dbKeyRefcountsEnabled() []byte {
	return []byte{partitionRefcountsEnabled}
}

func (tr *TrieR) fetchNodeData(nodeCommitment Hash) (*NodeData, bool) {
	nodeBin := tr.store.Get(dbKeyNodeData(nodeCommitment))
	if len(nodeBin) == 0 {
		return nil, false
	}
	ret, err := nodeDataFromBytes(nodeBin)
	assertf(err == nil, "NodeStore::FetchNodeData err: '%v' nodeBin: '%s', commitment: %s",
		err, hex.EncodeToString(nodeBin), nodeCommitment)
	ret.Commitment = nodeCommitment
	return ret, true
}

func (tr TrieR) fetchValueOfTerminal(terminal *Tcommitment) []byte {
	value, inTheCommitment := terminal.ExtractValue()
	if !inTheCommitment {
		value = tr.store.Get(terminal.dbKeyValue())
	}
	assertf(len(value) > 0, "can't fetch value. data commitment: %s", terminal)
	return value
}

// fetchChildrenNodeData fetches the children nodes of a given node in a single call to MultiGet.
func (tr TrieR) fetchChildrenNodeData(n *NodeData) [NumChildren]*NodeData {
	// fetch using a single call to MultiGet
	dbKeys := make([][]byte, 0, NumChildren)
	for _, hash := range n.Children {
		if hash == nil {
			continue
		}
		dbKeys = append(dbKeys, dbKeyNodeData(*hash))
	}

	children := [NumChildren]*NodeData{}
	if len(dbKeys) == 0 {
		// nothing to fetch
		return children
	}

	nodeBins := tr.store.MultiGet(dbKeys)

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
func (tr TrieR) fetchChildrenNodeDataWithRefcounts(n *NodeData) [NumChildren]NodeDataWithRefcounts {
	// fetch nodes with MultiGet
	nodeDatas := tr.fetchChildrenNodeData(n)
	// now fetch their refcounts with another MultiGet
	dbKeys := make([][]byte, 0, NumChildren*2)
	for _, child := range nodeDatas {
		if child == nil {
			continue
		}
		dbKeys = append(dbKeys, child.dbKeyNodeRefcount())
		if child.CommitsToExternalValue() {
			dbKeys = append(dbKeys, child.Terminal.dbKeyValueRefcount())
		}
	}
	var ret [NumChildren]NodeDataWithRefcounts
	if len(dbKeys) == 0 {
		// nothing to fetch
		return ret
	}
	refCountBytes := tr.store.MultiGet(dbKeys)
	for i, child := range nodeDatas {
		if child == nil {
			continue
		}
		nodeRefcount := codec.MustDecode[uint32](refCountBytes[0], 0)
		refCountBytes = refCountBytes[1:]
		valueRefcount := uint32(0)
		if child.CommitsToExternalValue() {
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

func (tr *TrieR) mustFetchNodeData(nodeCommitment Hash) *NodeData {
	ret, ok := tr.fetchNodeData(nodeCommitment)
	assertf(ok, "NodeStore::MustFetchNodeData: cannot find node data: commitment: '%s'", nodeCommitment.String())
	return ret
}

func (tr *TrieR) fetchChild(n *NodeData, childIdx byte, trieKey []byte) (*NodeData, []byte) {
	c := n.Children[childIdx]
	if c == nil {
		return nil, nil
	}
	childTriePath := concat(trieKey, n.PathExtension, []byte{childIdx})

	ret, ok := tr.fetchNodeData(*c)
	assertf(ok, "immutable::FetchChild: failed to fetch node. trieKey: '%s', childIndex: %d",
		hex.EncodeToString(trieKey), childIdx)
	return ret, childTriePath
}
