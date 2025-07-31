package trie

import (
	"io"

	"github.com/iotaledger/wasp/v2/packages/util/rwutil"
)

func (tr *TrieReader) TakeSnapshot(w io.Writer) error {
	// Some duplicated nodes and values might be written more than once in the snapshot;
	// Using a size-capped map to prevent this.
	// If the cap is reached, the generated snapshot will contain duplicate information,
	// but will still be correct.
	seenNodes := make(map[Hash]struct{})
	seenValues := make(map[string]struct{})
	const mapSizeCap = 2_000_000 / HashSizeBytes // 2 MB max for each map

	ww := rwutil.NewWriter(w)
	tr.IterateNodes(func(_ []byte, n *NodeData, depth int) IterateNodesAction {
		if _, seen := seenNodes[n.Commitment]; seen {
			return IterateContinue
		}
		if len(seenNodes) < mapSizeCap {
			seenNodes[n.Commitment] = struct{}{}
		}

		ww.WriteBytes(n.Bytes())
		if n.Terminal != nil && !n.Terminal.IsValue {
			valueKey := n.Terminal.Bytes()
			if _, seen := seenValues[string(valueKey)]; !seen {
				ww.WriteBool(true)
				value := tr.nodeStore.valueStore.Get(valueKey)
				ww.WriteBytes(value)
				if len(seenValues) < mapSizeCap {
					seenValues[string(valueKey)] = struct{}{}
				}
			} else {
				ww.WriteBool(false)
			}
		}
		if ww.Err != nil {
			return IterateStop
		}
		return IterateContinue
	})
	return ww.Err
}

func RestoreSnapshot(r io.Reader, store KVStore, refcountsEnabled bool) error {
	err := UpdateRefcountsFlag(store, refcountsEnabled)
	if err != nil {
		return err
	}

	triePartition := makeWriterPartition(store, partitionTrieNodes)
	valuePartition := makeWriterPartition(store, partitionValues)
	rr := rwutil.NewReader(r)
	var trieRoot *Hash
	for rr.Err == nil {
		nodeBytes := rr.ReadBytes()
		if rr.Err == io.EOF {
			break
		}
		if rr.Err != nil {
			return rr.Err
		}
		n, err := nodeDataFromBytes(nodeBytes)
		if err != nil {
			return err
		}
		n.updateCommitment()

		triePartition.Set(n.Commitment.Bytes(), nodeBytes)
		if trieRoot == nil {
			trieRoot = &n.Commitment
		}
		if n.Terminal != nil && !n.Terminal.IsValue && rr.ReadBool() {
			value := rr.ReadBytes()
			if rr.Err != nil {
				return rr.Err
			}
			valueKey := n.Terminal.Bytes()
			valuePartition.Set(valueKey, value)
		}
	}

	// TODO: improve performance
	if refcountsEnabled {
		_, refcounts := NewRefcounts(store)
		tr, err := NewTrieReader(store, *trieRoot)
		if err != nil {
			return err
		}
		tr.IterateNodes(func(nodeKey []byte, n *NodeData, depth int) IterateNodesAction {
			nodeRefcount := refcounts.GetNode(n.Commitment)
			nodeRefcount++
			refcounts.SetNode(n.Commitment, nodeRefcount)
			if nodeRefcount > 1 {
				return IterateSkipSubtree
			}
			if n.Terminal != nil && !n.Terminal.IsValue {
				valueRefcount := refcounts.GetValue(n.Terminal.Data)
				valueRefcount++
				refcounts.SetValue(n.Terminal.Data, valueRefcount)
			}
			return IterateContinue
		})
	}
	return nil
}
