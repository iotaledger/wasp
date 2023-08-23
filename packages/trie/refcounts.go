package trie

import (
	"fmt"

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

func (r *Refcounts) GetNode(commitment Hash) uint32 {
	return getRefcount(r.nodes, commitment[:])
}

func (r *Refcounts) Inc(node *bufferedNode) (uint32, uint32) {
	nodeCount, valueCount := r.incNodeAndValue(node.nodeData)
	node.nodeData.iterateChildren(func(i byte, commitment Hash) bool {
		if _, ok := node.uncommittedChildren[i]; !ok {
			// the new node adds a reference to an "old" node
			n := r.incNode(commitment)
			assertf(n > 1, "inconsistency")
		}
		return true
	})
	return nodeCount, valueCount
}

func (r *Refcounts) incNodeAndValue(node *NodeData) (uint32, uint32) {
	n := r.incNode(node.Commitment)
	v := uint32(0)
	if n == 1 && node.Terminal != nil && !node.Terminal.IsValue {
		v = incRefcount(r.values, node.Terminal.Data)
	}
	return n, v
}

func (r *Refcounts) incNode(commitment Hash) uint32 {
	return incRefcount(r.nodes, commitment[:])
}

func (r *Refcounts) Dec(node *NodeData) (deleteNode, deleteValue bool) {
	n := decRefcount(r.nodes, node.Commitment[:])
	deleteNode = n == 0
	if n == 0 && node.Terminal != nil && !node.Terminal.IsValue {
		nv := decRefcount(r.values, node.Terminal.Data)
		deleteValue = nv == 0
	}
	return
}

func (r *Refcounts) DebugDump() {
	fmt.Print("[node refcounts]\n")
	r.nodes.IterateKeys(func(k []byte) bool {
		n := getRefcount(r.nodes, k)
		fmt.Printf("   %x: %d\n", k, n)
		return true
	})
	fmt.Print("[value refcounts]\n")
	r.values.IterateKeys(func(k []byte) bool {
		n := getRefcount(r.values, k)
		fmt.Printf("   %x: %d\n", k, n)
		return true
	})
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
