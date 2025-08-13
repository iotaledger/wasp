package trie

import (
	"fmt"
	"io"

	"github.com/samber/lo"

	"github.com/iotaledger/wasp/v2/packages/kv/codec"
)

type Refcounts struct {
	store KVStore
}

func initRefcounts(store KVStore, root *draftNode) {
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

// IsRefcountsEnabled reads the enabled flag from the db.
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

// DeleteRefcountsFlag deletes the refcounts enabled flag from the store.
// This is useful for testing purposes, to reset the state of the store.
func DeleteRefcountsFlag(store KVStore) {
	store.Del([]byte{partitionRefcountsEnabled})
}

func (r *Refcounts) GetNode(commitment Hash) uint32 {
	return r.getRefcount(nodeRefcountKey(commitment[:]))
}

func (r *Refcounts) GetValue(commitment []byte) uint32 {
	return r.getRefcount(valueRefcountKey(commitment))
}

// inc is called after a commit operation, and increments the refcounts for all affected nodes
func (r *Refcounts) inc(root *draftNode) CommitStats {
	nodeRefcounts, valueRefcounts := r.fetchBeforeCommit(root)

	// increment and write updated refcounts
	//
	// writing updated values in a batch is not necessary here because it is
	// already handled by the underlying store
	stats := CommitStats{}
	incrementNode := func(commitment Hash) uint32 {
		refcount := nodeRefcounts[commitment]
		refcount++
		nodeRefcounts[commitment] = refcount
		r.setRefcount(nodeRefcountKey(commitment[:]), refcount)
		if refcount == 1 {
			stats.CreatedNodes++
		}
		return refcount
	}
	incrementValue := func(data []byte) uint32 {
		refcount := valueRefcounts[string(data)]
		refcount++
		valueRefcounts[string(data)] = refcount
		r.setRefcount(valueRefcountKey(data), refcount)
		if refcount == 1 {
			stats.CreatedValues++
		}
		return refcount
	}
	root.traversePreOrder(func(node *draftNode) IterateNodesAction {
		nodeRefcount := incrementNode(node.nodeData.Commitment)
		if nodeRefcount > 1 {
			// don't increment its children refcounts
			return IterateSkipSubtree
		}
		// this is a new node, increment its children refcounts
		if node.terminal != nil && !node.terminal.IsValue {
			_ = incrementValue(node.terminal.Data)
		}
		node.nodeData.iterateChildren(func(i byte, childCommitment Hash) bool {
			if _, ok := node.uncommittedChildren[i]; !ok {
				// a new node adds a reference to an old node
				childRefcount := incrementNode(childCommitment)
				assertf(childRefcount > 1, "inconsistency %s %s %d", node.nodeData.Commitment, childCommitment, childRefcount)
			}
			return true
		})
		return IterateContinue
	})
	return stats
}

// fetchBeforeCommit fetches all affected refcounts, using a single call to
// MultiGet
func (r *Refcounts) fetchBeforeCommit(root *draftNode) (
	nodeRefcounts map[Hash]uint32,
	valueRefcounts map[string]uint32,
) {
	type fetch struct {
		isNode         bool
		dbKey          []byte
		nodeCommitment Hash
		valueData      []byte
	}
	var toFetch []fetch

	addNode := func(commitment Hash) {
		toFetch = append(toFetch, fetch{
			isNode:         true,
			dbKey:          nodeRefcountKey(commitment[:]),
			nodeCommitment: commitment,
		})
	}
	addValue := func(nodeData []byte) {
		toFetch = append(toFetch, fetch{
			isNode:    false,
			dbKey:     valueRefcountKey(nodeData),
			valueData: nodeData,
		})
	}
	visited := make(map[Hash]struct{})
	root.traversePreOrder(func(node *draftNode) IterateNodesAction {
		if _, ok := visited[node.nodeData.Commitment]; ok {
			// this node was already visited, don't fetch twice
			return IterateSkipSubtree
		}
		visited[node.nodeData.Commitment] = struct{}{}

		addNode(node.nodeData.Commitment)
		if node.terminal != nil && !node.terminal.IsValue {
			addValue(node.terminal.Data)
		}
		// also fetch old nodes referenced by this node
		node.nodeData.iterateChildren(func(i byte, childCommitment Hash) bool {
			if _, ok := node.uncommittedChildren[i]; !ok {
				addNode(childCommitment)
			}
			return true
		})
		return IterateContinue
	})
	refCountBytes := r.store.MultiGet(lo.Map(toFetch, func(f fetch, _ int) []byte {
		return f.dbKey
	}))
	nodeRefcounts = make(map[Hash]uint32)
	valueRefcounts = make(map[string]uint32)
	for i, f := range toFetch {
		if f.isNode {
			nodeRefcounts[f.nodeCommitment] = codec.MustDecode[uint32](refCountBytes[i], 0)
		} else {
			valueRefcounts[string(f.valueData)] = codec.MustDecode[uint32](refCountBytes[i], 0)
		}
	}
	return nodeRefcounts, valueRefcounts
}

func (r *Refcounts) SetNode(hash Hash, n uint32) {
	r.setRefcount(nodeRefcountKey(hash[:]), n)
}

func (r *Refcounts) SetValue(key []byte, n uint32) {
	r.setRefcount(valueRefcountKey(key), n)
}

func (r *Refcounts) DebugDump(w io.Writer) (
	nodeRefcounts map[Hash]uint32,
	valueRefcounts map[string]uint32,
) {
	nodeRefcounts = make(map[Hash]uint32)
	valueRefcounts = make(map[string]uint32)

	fmt.Fprint(w, "[node refcounts]\n")
	makeKVStorePartition(r.store, partitionRefcountNodes).IterateKeys(func(k []byte) bool {
		n := r.getRefcount(nodeRefcountKey(k))
		fmt.Fprintf(w, "   %x: %d\n", k, n)
		nodeRefcounts[Hash(k)] = n
		return true
	})
	fmt.Fprint(w, "[value refcounts]\n")
	makeKVStorePartition(r.store, partitionRefcountValues).IterateKeys(func(k []byte) bool {
		n := r.getRefcount(valueRefcountKey(k))
		fmt.Fprintf(w, "   %x: %d\n", k, n)
		valueRefcounts[string(k)] = n
		return true
	})
	return
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
