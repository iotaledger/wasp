package trie

import (
	"errors"
)

type PruneStats struct {
	DeletedNodes  uint
	DeletedValues uint
}

// Prune decrements the refcount of the trie root and all its children,
// and then deletes all nodes and values that have a refcount of 0.
func (tr *TrieRW) Prune(trieRoot Hash) (PruneStats, error) {
	if !tr.IsRefcountsEnabled() {
		return PruneStats{}, errors.New("refcounts disabled, cannot prune trie")
	}

	touchedRefcounts := NewRefcounts()

	// decrement refcounts
	NewTrieRFromRoot(tr.store, trieRoot).IterateNodesWithRefcounts(func(nodeKey []byte, n *NodeData, depth int, nr, vr uint32) IterateNodesAction {
		nodeRefcount := touchedRefcounts.setNodeIfAbsent(n.Commitment, nr)
		if nodeRefcount == 0 {
			// node already deleted
			return IterateSkipSubtree
		}
		nodeRefcount = touchedRefcounts.decNode(n.Commitment)
		if nodeRefcount == 0 && n.CommitsToExternalValue() {
			valueRefcount := touchedRefcounts.setValueIfAbsent(n.Terminal, vr)
			if valueRefcount > 0 {
				touchedRefcounts.decValue(n.Terminal)
			}
		}
		if nodeRefcount == 0 {
			// node deleted => decrease refcount of children
			return IterateContinue
		}
		// node not deleted => do not decrease refcount of children
		return IterateSkipSubtree
	})

	// write modified refcounts and delete nodes/values with refcount 0
	stats := PruneStats{}
	for hash, n := range touchedRefcounts.Nodes {
		tr.setNodeRefcount(hash, n)
		if n == 0 {
			tr.store.Del(dbKeyNodeData(hash))
			stats.DeletedNodes++
		}
	}
	for terminalData, n := range touchedRefcounts.Values {
		t := &Tcommitment{Data: []byte(terminalData), IsValue: false}
		tr.setValueRefcount(t, n)
		if n == 0 {
			tr.store.Del(t.dbKeyValue())
			stats.DeletedValues++
		}
	}
	return stats, nil
}
