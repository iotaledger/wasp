package trie

import (
	"errors"

	"github.com/samber/lo"

	"github.com/iotaledger/wasp/v2/packages/util/rwutil"
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

	touchedNodes := make(map[Hash]uint32)
	touchedValues := make(map[string]uint32)

	// decrement refcounts
	NewTrieRFromRoot(tr.store, trieRoot).IterateNodesWithRefcounts(func(nodeKey []byte, n *NodeData, depth int, nr, vr uint32) IterateNodesAction {
		nodeRefcount := lo.ValueOr(touchedNodes, n.Commitment, nr)
		if nodeRefcount == 0 {
			// node already deleted
			return IterateSkipSubtree
		}
		nodeRefcount--
		touchedNodes[n.Commitment] = nodeRefcount
		if nodeRefcount == 0 && n.CommitsToExternalValue() {
			valueBytes := string(n.Terminal.Bytes())
			valueRefcount := lo.ValueOr(touchedValues, valueBytes, vr)
			if valueRefcount > 0 {
				touchedValues[valueBytes] = valueRefcount - 1
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
	for hash, n := range touchedNodes {
		tr.setNodeRefcount(hash, n)
		if n == 0 {
			tr.store.Del(dbKeyNodeData(hash))
			stats.DeletedNodes++
		}
	}
	for valueBytes, n := range touchedValues {
		t := lo.Must(rwutil.ReadFromBytes([]byte(valueBytes), &Tcommitment{}))
		tr.setValueRefcount(t.Data, n)
		if n == 0 {
			tr.store.Del(dbKeyValue([]byte(valueBytes)))
			stats.DeletedValues++
		}
	}
	return stats, nil
}
