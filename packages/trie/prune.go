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
func Prune(store KVStore, trieRoot Hash) (PruneStats, error) {
	enabled, refcounts := NewRefcounts(store)
	if !enabled {
		return PruneStats{}, errors.New("refcounts disabled, cannot prune trie")
	}

	tr, err := NewTrieReader(store, trieRoot)
	if err != nil {
		return PruneStats{}, err
	}

	touchedNodes := make(map[Hash]uint32)
	touchedValues := make(map[string]uint32)

	// decrement refcounts
	// TODO: modify IterateNodes so that children refcounts are fetched using MultiGet
	tr.IterateNodes(func(nodeKey []byte, n *NodeData, depth int) IterateNodesAction {
		nodeRefcount := lo.ValueOr(touchedNodes, n.Commitment, refcounts.GetNode(n.Commitment))
		if nodeRefcount == 0 {
			// node already deleted
			return IterateSkipSubtree
		}
		nodeRefcount--
		touchedNodes[n.Commitment] = nodeRefcount
		if nodeRefcount == 0 && n.Terminal != nil && !n.Terminal.IsValue {
			valueBytes := string(n.Terminal.Bytes())
			valueRefcount := lo.ValueOr(touchedValues, valueBytes, refcounts.GetValue(n.Terminal.Data))
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
	triePartition := makeWriterPartition(store, partitionTrieNodes)
	stats := PruneStats{}
	for hash, n := range touchedNodes {
		refcounts.SetNode(hash, n)
		if n == 0 {
			triePartition.Del(hash[:])
			stats.DeletedNodes++
		}
	}
	valuePartition := makeWriterPartition(store, partitionValues)
	for valueBytes, n := range touchedValues {
		t := lo.Must(rwutil.ReadFromBytes([]byte(valueBytes), &Tcommitment{}))
		refcounts.SetValue(t.Data, n)
		if n == 0 {
			valuePartition.Del([]byte(valueBytes))
			stats.DeletedValues++
		}
	}
	return stats, nil
}
