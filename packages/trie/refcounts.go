package trie

import (
	"fmt"
	"io"

	"github.com/samber/lo"

	"github.com/iotaledger/wasp/v2/packages/kv/codec"
)

func (tr *TrieRW) initRefcounts(root *draftNode) {
	enabled := tr.IsRefcountsEnabled()
	if enabled {
		tr.incRefcounts(root)
	}
}

// IsRefcountsEnabled reads the enabled flag from the db.
// It returns true only after EnableRefcounts is called on an empty DB.
func (tr *TrieR) IsRefcountsEnabled() bool {
	return lo.Must(codec.Decode[bool](tr.store.Get(dbKeyRefcountsEnabled()), false))
}

// UpdateRefcountsFlag enables or disables trie reference counting.
// Note that the flag can only be set to true on an empty DB.
func (tr *TrieRW) UpdateRefcountsFlag(enable bool) error {
	if !enable {
		// flag can be disabled without restrictions
		tr.store.Set([]byte{partitionRefcountsEnabled}, codec.Encode(false))
		return nil
	}
	enabled := tr.IsRefcountsEnabled()
	if enabled {
		// already enabled
		return nil
	}
	// flag can be enabled only on an empty store
	isEmpty := true
	tr.store.IterateKeys(nil, func(k []byte) bool {
		isEmpty = false
		return false
	})
	if !isEmpty {
		return fmt.Errorf("cannot enable refcounts on a non-empty store")
	}

	tr.store.Set([]byte{partitionRefcountsEnabled}, codec.Encode(true))
	return nil
}

// DeleteRefcountsFlag deletes the refcounts enabled flag from the store.
// This is useful for testing purposes, to reset the state of the store.
func DeleteRefcountsFlag(store KVStore) {
	store.Del([]byte{partitionRefcountsEnabled})
}

func (tr *TrieR) GetNodeRefcount(commitment Hash) uint32 {
	return tr.getRefcount(dbKeyNodeRefcount(commitment[:]))
}

func (tr *TrieRW) setNodeRefcount(commitment Hash, n uint32) {
	tr.setRefcount(dbKeyNodeRefcount(commitment[:]), n)
}

func (tr *TrieR) GetValueRefcount(terminalData []byte) uint32 {
	return tr.getRefcount(dbKeyValueRefcount(terminalData))
}

func (tr *TrieRW) setValueRefcount(terminalData []byte, n uint32) {
	tr.setRefcount(dbKeyValueRefcount(terminalData), n)
}

// incRefcounts is called after a commit operation, and increments the refcounts for all affected nodes
func (tr *TrieRW) incRefcounts(root *draftNode) CommitStats {
	nodeRefcounts, valueRefcounts := tr.fetchBeforeCommit(root)

	// increment and write updated refcounts
	//
	// writing updated values in a batch is not necessary here because it is
	// already handled by the underlying store
	stats := CommitStats{}
	incrementNode := func(commitment Hash) uint32 {
		refcount := nodeRefcounts[commitment]
		refcount++
		nodeRefcounts[commitment] = refcount
		tr.setNodeRefcount(commitment, refcount)
		if refcount == 1 {
			stats.CreatedNodes++
		}
		return refcount
	}
	incrementValue := func(terminalData []byte) uint32 {
		refcount := valueRefcounts[string(terminalData)]
		refcount++
		valueRefcounts[string(terminalData)] = refcount
		tr.setValueRefcount(terminalData, refcount)
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
		if node.CommitsToExternalValue() {
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
func (tr *TrieR) fetchBeforeCommit(root *draftNode) (
	nodeRefcounts map[Hash]uint32,
	valueRefcounts map[string]uint32,
) {
	type fetch struct {
		isNode         bool
		dbKey          []byte
		nodeCommitment Hash
		terminalData   []byte
	}
	var toFetch []fetch

	addNode := func(commitment Hash) {
		toFetch = append(toFetch, fetch{
			isNode:         true,
			dbKey:          dbKeyNodeRefcount(commitment[:]),
			nodeCommitment: commitment,
		})
	}
	addValue := func(terminalData []byte) {
		toFetch = append(toFetch, fetch{
			isNode:       false,
			dbKey:        dbKeyValueRefcount(terminalData),
			terminalData: terminalData,
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
		if node.CommitsToExternalValue() {
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
	refcountBytes := tr.store.MultiGet(lo.Map(toFetch, func(f fetch, _ int) []byte {
		return f.dbKey
	}))
	nodeRefcounts = make(map[Hash]uint32)
	valueRefcounts = make(map[string]uint32)
	for i, f := range toFetch {
		if f.isNode {
			nodeRefcounts[f.nodeCommitment] = codec.MustDecode[uint32](refcountBytes[i], 0)
		} else {
			valueRefcounts[string(f.terminalData)] = codec.MustDecode[uint32](refcountBytes[i], 0)
		}
	}
	return nodeRefcounts, valueRefcounts
}

func (tr *TrieR) DebugDumpRefcounts(w io.Writer) (
	nodeRefcounts map[Hash]uint32,
	valueRefcounts map[string]uint32,
) {
	nodeRefcounts = make(map[Hash]uint32)
	valueRefcounts = make(map[string]uint32)

	fmt.Fprint(w, "[node refcounts]\n")
	tr.store.IterateKeys([]byte{partitionRefcountNodes}, func(k []byte) bool {
		commitment, err := HashFromBytes(k[1:])
		assertNoError(err)
		n := tr.GetNodeRefcount(commitment)
		fmt.Fprintf(w, "   %x: %d\n", k, n)
		nodeRefcounts[commitment] = n
		return true
	})
	fmt.Fprint(w, "[value refcounts]\n")
	tr.store.IterateKeys([]byte{partitionRefcountValues}, func(k []byte) bool {
		terminalData := k[1:]
		n := tr.GetValueRefcount(terminalData)
		fmt.Fprintf(w, "   %x: %d\n", k, n)
		valueRefcounts[string(terminalData)] = n
		return true
	})
	return
}

func (tr *TrieR) getRefcount(key []byte) uint32 {
	b := tr.store.Get(key)
	if b == nil {
		return 0
	}
	return codec.MustDecode[uint32](b)
}

func (tr *TrieRW) setRefcount(key []byte, n uint32) {
	if n == 0 {
		tr.store.Del(key)
	} else {
		tr.store.Set(key, codec.Encode[uint32](n))
	}
}
