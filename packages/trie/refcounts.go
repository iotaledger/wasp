package trie

import (
	"github.com/iotaledger/wasp/packages/kv/codec"
)

type Refcounts struct {
	nodes  KVStore
	values KVStore
}

func newRefcounts(store KVStore) *Refcounts {
	return &Refcounts{
		nodes:  makeKVStorePartition(store, partitionRefcountNodes),
		values: makeKVStorePartition(store, partitionRefcountValues),
	}
}

func (d *Refcounts) Inc(node *bufferedNode) (uint32, uint32) {
	nodeCount, valueCount := d.incNodeAndValue(node.nodeData)
	node.nodeData.iterateChildren(func(i byte, commitment Hash) bool {
		if _, ok := node.uncommittedChildren[i]; !ok {
			// the new node adds a reference to an "old" node
			n := d.incNode(commitment)
			assertf(n > 1, "inconsistency")
		}
		return true
	})
	return nodeCount, valueCount
}

func (d *Refcounts) incNodeAndValue(node *NodeData) (uint32, uint32) {
	n := d.incNode(node.Commitment)
	v := uint32(0)
	if n == 1 && node.Terminal != nil && !node.Terminal.IsValue {
		v = incRefcount(d.values, node.Terminal.Data)
	}
	return n, v
}

func (d *Refcounts) incNode(commitment Hash) uint32 {
	return incRefcount(d.nodes, commitment[:])
}

func (d *Refcounts) Dec(node *NodeData) (deleteNode, deleteValue bool) {
	n := decRefcount(d.nodes, node.Commitment[:])
	deleteNode = n == 0
	if n == 0 && node.Terminal != nil && !node.Terminal.IsValue {
		nv := decRefcount(d.values, node.Terminal.Data)
		deleteValue = nv == 0
	}
	return
}

func incRefcount(s KVStore, key []byte) uint32 {
	n := getRefcount(s, key)
	n++
	setRefcount(s, key, n)
	return n
}

func decRefcount(s KVStore, key []byte) uint32 {
	n := getRefcount(s, key)
	if n == 0 {
		panic("inconsistency: negative refcount")
	}
	n--
	setRefcount(s, key, n)
	return n
}

func getRefcount(s KVStore, key []byte) uint32 {
	b := s.Get(key)
	if b == nil {
		return 0
	}
	return codec.MustDecodeUint32(b)
}

func setRefcount(s KVStore, key []byte, n uint32) {
	if n == 0 {
		s.Del(key)
	} else {
		s.Set(key, codec.EncodeUint32(n))
	}
}
