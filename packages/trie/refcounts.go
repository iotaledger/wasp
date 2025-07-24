package trie

import (
	"fmt"

	"github.com/samber/lo"

	"github.com/iotaledger/wasp/v2/packages/kv/codec"
)

type Refcounts struct {
	store KVStore
}

func initRefcounts(store KVStore, root *bufferedNode) {
	enabled, refcounts := NewRefcounts(store)
	if enabled {
		refcounts.inc(root)
	}
}

func NewRefcounts(store KVStore) (bool, *Refcounts) {
	enabled := IsRefcountsEnabled(store)
	if !enabled {
		return false, nil
	}
	return true, &Refcounts{store: store}
}

// isRefcountsEnabled reads the enabled flag from the db.
// It returns true only after EnableRefcounts is called on an empty DB.
func IsRefcountsEnabled(store KVStore) bool {
	return lo.Must(codec.Decode[bool](store.Get([]byte{partitionRefcountsEnabled}), false))
}

// UpdateRefcountsFlag enables or disables trie reference counting.
// Note that the flag can only be set to true on an empty DB.
func UpdateRefcountsFlag(store KVStore, enable bool) error {
	if !enable {
		// flag can be disabled without restrictions
		store.Set([]byte{partitionRefcountsEnabled}, codec.Encode(false))
		return nil
	}
	enabled := IsRefcountsEnabled(store)
	if enabled {
		// already enabled
		return nil
	}
	// flag can be enabled only on an empty store
	isEmpty := true
	store.IterateKeys(func(k []byte) bool {
		isEmpty = false
		return false
	})
	if !isEmpty {
		return fmt.Errorf("cannot enable refcounts on a non-empty store")
	}

	store.Set([]byte{partitionRefcountsEnabled}, codec.Encode(true))
	return nil
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

	visited := make(map[Hash]struct{})
	root.traversePreOrder(func(node *bufferedNode) IterateNodesAction {
		if _, ok := visited[node.nodeData.Commitment]; ok {
			// this node was already visited, don't fetch twice
			return IterateSkipSubtree
		}
		visited[node.nodeData.Commitment] = struct{}{}

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
		return IterateContinue
	})
	refCountBytes := r.store.MultiGet(dbKeys)
	refCounts := lo.Map(refCountBytes, func(v []byte, _ int) uint32 {
		if v == nil {
			return 0
		}
		return codec.MustDecode[uint32](v)
	})
	// treat the refCounts slice as a queue
	nextRefcout := func() (r uint32) {
		r, refCounts = refCounts[0], refCounts[1:]
		return r
	}

	// increment and write updated refcounts
	//
	// writing updated values in a batch is not necessary here because it is
	// already handled by the underlying store
	touchedRefcounts := make(map[string]uint32)
	incrementAndSet := func(refcount uint32, dbKey []byte, createdCount *uint) uint32 {
		refcount = lo.ValueOr(touchedRefcounts, string(dbKey), refcount)
		refcount++
		touchedRefcounts[string(dbKey)] = refcount
		r.setRefcount(dbKey, refcount)
		if createdCount != nil && refcount == 1 {
			*createdCount++
		}
		return refcount
	}
	stats := CommitStats{}
	visited = make(map[Hash]struct{})
	root.traversePreOrder(func(node *bufferedNode) IterateNodesAction {
		if _, ok := visited[node.nodeData.Commitment]; ok {
			// this node was already visited, we only want to incremtnt its
			// refcount but not its children's
			dbKey := nodeRefcountKey(node.nodeData.Commitment[:])
			nodeRefcount := touchedRefcounts[string(dbKey)]
			assertf(nodeRefcount > 0, "inconsistency %s %d", node.nodeData.Commitment, nodeRefcount)
			incrementAndSet(nodeRefcount, dbKey, &stats.CreatedNodes)
			return IterateSkipSubtree
		}
		visited[node.nodeData.Commitment] = struct{}{}

		nodeRefcount := incrementAndSet(nextRefcout(), nodeRefcountKey(node.nodeData.Commitment[:]), &stats.CreatedNodes)

		if node.terminal != nil && !node.terminal.IsValue {
			valueRefcount := nextRefcout()
			if nodeRefcount == 1 {
				// a new node adds a reference to a value
				valueRefcount = incrementAndSet(valueRefcount, valueRefcountKey(node.terminal.Data), &stats.CreatedValues)
			}
		}

		node.nodeData.iterateChildren(func(i byte, childCommitment Hash) bool {
			if _, ok := node.uncommittedChildren[i]; !ok {
				childRefcount := nextRefcout()
				if nodeRefcount == 1 {
					// a new node adds a reference to an old node
					assertf(childRefcount > 0, "inconsistency %s %s %d", node.nodeData.Commitment, childCommitment, childRefcount)
					childRefcount = incrementAndSet(childRefcount, nodeRefcountKey(childCommitment[:]), nil)
				}
			}
			return true
		})

		return IterateContinue
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
