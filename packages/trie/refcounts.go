package trie

import (
	"fmt"

	"github.com/iotaledger/wasp/v2/packages/kv/codec"
)

type Refcounts struct {
	store KVStore
}

func NewRefcounts(store KVStore) *Refcounts {
	return newRefcounts(store)
}

func newRefcounts(store KVStore) *Refcounts {
	return &Refcounts{store: store}
}

func (r *Refcounts) GetNode(commitment Hash) uint32 {
	return r.getRefcount(nodeRefcountKey(commitment[:]))
}

func (r *Refcounts) GetValue(commitment []byte) uint32 {
	return r.getRefcount(valueRefcountKey(commitment))
}

// inc is called after a commit operation, and increments the refcounts for all affected nodes
func (r *Refcounts) inc(root *bufferedNode) CommitStats {
	// use MultiGet to read all refcounts that are affected
	var dbKeys [][]byte
	root.traversePreOrder(func(node *bufferedNode) {
		dbKeys = append(dbKeys, nodeRefcountKey(node.nodeData.Commitment[:]))
		if node.terminal != nil && !node.terminal.IsValue {
			dbKeys = append(dbKeys, valueRefcountKey(node.terminal.Data))
		}
		// also fetch old nodes referenced by this node
		node.nodeData.iterateChildren(func(i byte, childCommitment Hash) bool {
			if _, ok := node.uncommittedChildren[i]; !ok {
				dbKeys = append(dbKeys, nodeRefcountKey(childCommitment[:]))
			}
			return true
		})
	})
	refCountBytes := r.store.MultiGet(dbKeys)
	refCounts := lo.Map(refCountBytes, func(v []byte, _ int) uint32 {
		if v == nil {
			return 0
		}
		return codec.MustDecode[uint32](v)
	})

	// increment and write updated refcounts
	//
	// writing updated values in a batch is not necessary here because it is
	// already handled by the underlying store
	incrementAndSet := func(refcount uint32, dbKey []byte, createdCount *uint) uint32 {
		refcount++
		r.setRefcount(dbKey, refcount)
		if createdCount != nil && refcount == 1 {
			*createdCount++
		}
		return refcount
	}
	stats := CommitStats{}
	root.traversePreOrder(func(node *bufferedNode) {
		var nodeRefcount uint32
		nodeRefcount, refCounts = refCounts[0], refCounts[1:]
		nodeRefcount = incrementAndSet(nodeRefcount, nodeRefcountKey(node.nodeData.Commitment[:]), &stats.CreatedNodes)

		if node.terminal != nil && !node.terminal.IsValue {
			var valueRefcount uint32
			valueRefcount, refCounts = refCounts[0], refCounts[1:]
			if nodeRefcount == 1 {
				// a new node adds a reference to a value
				valueRefcount = incrementAndSet(valueRefcount, valueRefcountKey(node.terminal.Data), &stats.CreatedValues)
			}
		}

		node.nodeData.iterateChildren(func(i byte, childCommitment Hash) bool {
			if _, ok := node.uncommittedChildren[i]; !ok {
				var childRefcount uint32
				childRefcount, refCounts = refCounts[0], refCounts[1:]
				if nodeRefcount == 1 {
					// a new node adds a reference to an old node
					assertf(childRefcount > 0, "inconsistency %s %s %d", node.nodeData.Commitment, childCommitment, childRefcount)
					childRefcount = incrementAndSet(childRefcount, nodeRefcountKey(childCommitment[:]), nil)
				}
			}
			return true
		})
	})
	assertf(len(refCounts) == 0, "inconsistency: remaining refCounts")
	return stats
}

func (r *Refcounts) SetNode(hash Hash, n uint32) {
	r.setRefcount(nodeRefcountKey(hash[:]), n)
}

func (r *Refcounts) SetValue(key []byte, n uint32) {
	r.setRefcount(valueRefcountKey(key), n)
}

func (r *Refcounts) DebugDump() {
	fmt.Print("[node refcounts]\n")
	makeKVStorePartition(r.store, partitionRefcountNodes).IterateKeys(func(k []byte) bool {
		n := r.getRefcount(nodeRefcountKey(k))
		fmt.Printf("   %x: %d\n", k, n)
		return true
	})
	fmt.Print("[value refcounts]\n")
	makeKVStorePartition(r.store, partitionRefcountValues).IterateKeys(func(k []byte) bool {
		n := r.getRefcount(valueRefcountKey(k))
		fmt.Printf("   %x: %d\n", k, n)
		return true
	})
}

func nodeRefcountKey(nodeCommitment []byte) []byte {
	return append([]byte{partitionRefcountNodes}, nodeCommitment...)
}

func valueRefcountKey(valueCommitment []byte) []byte {
	return append([]byte{partitionRefcountValues}, valueCommitment...)
}

func (r *Refcounts) getRefcount(key []byte) uint32 {
	b := r.store.Get(key)
	if b == nil {
		return 0
	}
	return codec.MustDecode[uint32](b)
}

func (r *Refcounts) setRefcount(key []byte, n uint32) {
	if n == 0 {
		r.store.Del(key)
	} else {
		r.store.Set(key, codec.Encode[uint32](n))
	}
}
