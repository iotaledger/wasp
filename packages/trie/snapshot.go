package trie

import (
	"io"

	"github.com/iotaledger/wasp/packages/util/rwutil"
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

func RestoreSnapshot(r io.Reader, store KVStore) error {
	triePartition := makeWriterPartition(store, partitionTrieNodes)
	valuePartition := makeWriterPartition(store, partitionValues)
	refcounts := newRefcounts(store)
	rr := rwutil.NewReader(r)
	for {
		nodeBytes := rr.ReadBytes()
		if rr.Err == io.EOF {
			return nil
		}
		n, err := nodeDataFromBytes(nodeBytes)
		n.updateCommitment()
		if err != nil {
			return err
		}
		nodeKey := n.Commitment.Bytes()
		triePartition.Set(nodeKey, nodeBytes)
		if n.Terminal != nil && !n.Terminal.IsValue {
			if rr.ReadBool() {
				value := rr.ReadBytes()
				valueKey := n.Terminal.Bytes()
				valuePartition.Set(valueKey, value)
			}
		}
		refcounts.incNodeAndValue(n)
		if rr.Err != nil {
			return rr.Err
		}
	}
}
