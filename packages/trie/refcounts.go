package trie

import (
	"fmt"
	"io"

	"github.com/samber/lo"

	"github.com/iotaledger/wasp/v2/packages/kv/codec"
)

// Refcounts holds reference counts for nodes and values in memory
type Refcounts struct {
	Nodes  map[Hash]uint32
	Values map[string]uint32
}

func NewRefcounts() *Refcounts {
	return &Refcounts{
		Nodes:  make(map[Hash]uint32),
		Values: make(map[string]uint32),
	}
}

func incRefcount[T comparable](m map[T]uint32, k T) uint32 {
	m[k]++
	return m[k]
}

func decRefcount[T comparable](m map[T]uint32, k T) uint32 {
	n := m[k]
	if n == 0 {
		panic(fmt.Sprintf("cannot decrement refcount of %v, already 0", k))
	}
	n--
	m[k] = n
	return n
}

func setRefcountIfAbsent[T comparable](m map[T]uint32, k T, n uint32) uint32 {
	v, ok := m[k]
	if ok {
		return v
	}
	m[k] = n
	return n
}

func (r *Refcounts) incNode(commitment Hash) uint32 {
	return incRefcount(r.Nodes, commitment)
}

func (r *Refcounts) incValue(t *Tcommitment) uint32 {
	return incRefcount(r.Values, string(t.Data))
}

func (r *Refcounts) decNode(commitment Hash) uint32 {
	return decRefcount(r.Nodes, commitment)
}

func (r *Refcounts) decValue(t *Tcommitment) uint32 {
	return decRefcount(r.Values, string(t.Data))
}

func (r *Refcounts) setNodeIfAbsent(commitment Hash, n uint32) uint32 {
	return setRefcountIfAbsent(r.Nodes, commitment, n)
}

func (r *Refcounts) setValueIfAbsent(t *Tcommitment, n uint32) uint32 {
	return setRefcountIfAbsent(r.Values, string(t.Data), n)
}

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

func (tr *TrieR) GetValueRefcount(t *Tcommitment) uint32 {
	return tr.getRefcount(t.dbKeyValueRefcount())
}

func (tr *TrieRW) setValueRefcount(t *Tcommitment, n uint32) {
	tr.setRefcount(t.dbKeyValueRefcount(), n)
}

// incRefcounts is called after a commit operation, and increments the refcounts for all affected nodes
func (tr *TrieRW) incRefcounts(root *draftNode) CommitStats {
	refcounts := tr.fetchBeforeCommit(root)

	// increment and write updated refcounts
	//
	// writing updated values in a batch is not necessary here because it is
	// already handled by the underlying store
	stats := CommitStats{}
	incrementNode := func(commitment Hash) uint32 {
		refcount := refcounts.incNode(commitment)
		tr.setNodeRefcount(commitment, refcount)
		if refcount == 1 {
			stats.CreatedNodes++
		}
		return refcount
	}
	incrementValue := func(t *Tcommitment) uint32 {
		refcount := refcounts.incValue(t)
		tr.setValueRefcount(t, refcount)
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
			_ = incrementValue(node.terminal)
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
func (tr *TrieR) fetchBeforeCommit(root *draftNode) *Refcounts {
	type fetch struct {
		dbKey          []byte
		nodeCommitment Hash
		terminal       *Tcommitment
	}
	var toFetch []fetch

	addNode := func(commitment Hash) {
		toFetch = append(toFetch, fetch{
			dbKey:          dbKeyNodeRefcount(commitment[:]),
			nodeCommitment: commitment,
		})
	}
	addValue := func(t *Tcommitment) {
		toFetch = append(toFetch, fetch{
			dbKey:    t.dbKeyValueRefcount(),
			terminal: t,
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
			addValue(node.terminal)
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
	refcounts := NewRefcounts()
	for i, f := range toFetch {
		v := codec.MustDecode[uint32](refcountBytes[i], 0)
		if f.terminal == nil { // this is a node refcount
			refcounts.setNodeIfAbsent(f.nodeCommitment, v)
		} else { // this is a value refcount
			refcounts.setValueIfAbsent(f.terminal, v)
		}
	}
	return refcounts
}

func (tr *TrieR) DebugDumpRefcounts(w io.Writer) *Refcounts {
	refcounts := NewRefcounts()
	fmt.Fprint(w, "[node refcounts]\n")
	tr.store.IterateKeys([]byte{partitionRefcountNodes}, func(k []byte) bool {
		commitment, err := HashFromBytes(k[1:])
		assertNoError(err)
		n := tr.GetNodeRefcount(commitment)
		fmt.Fprintf(w, "   %x: %d\n", k, n)
		refcounts.setNodeIfAbsent(commitment, n)
		return true
	})
	fmt.Fprint(w, "[value refcounts]\n")
	tr.store.IterateKeys([]byte{partitionRefcountValues}, func(k []byte) bool {
		t := &Tcommitment{IsValue: false, Data: k[1:]}
		n := tr.GetValueRefcount(t)
		fmt.Fprintf(w, "   %x: %d\n", k, n)
		refcounts.setValueIfAbsent(t, n)
		return true
	})
	return refcounts
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
