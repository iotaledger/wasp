package trie

import (
	"io"

	"github.com/iotaledger/wasp/v2/packages/util/rwutil"
)

func (tr *TrieRFromRoot) TakeSnapshot(w io.Writer) error {
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
		if n.CommitsToExternalValue() {
			valueKey := n.Terminal.dbKeyValue()
			if _, seen := seenValues[string(valueKey)]; !seen {
				ww.WriteBool(true)
				value := tr.R.store.Get(valueKey)
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

func (tr *TrieRW) RestoreSnapshot(r io.Reader, refcountsEnabled bool) error {
	err := tr.UpdateRefcountsFlag(refcountsEnabled)
	if err != nil {
		return err
	}

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

		tr.store.Set(n.dbKey(), nodeBytes)
		if trieRoot == nil {
			trieRoot = &n.Commitment
		}
		if n.CommitsToExternalValue() && rr.ReadBool() {
			value := rr.ReadBytes()
			if rr.Err != nil {
				return rr.Err
			}
			tr.store.Set(n.Terminal.dbKeyValue(), value)
		}
	}

	if refcountsEnabled {
		touchedRefcounts := NewRefcounts()
		NewTrieRFromRoot(tr.store, *trieRoot).IterateNodesWithRefcounts(func(nodeKey []byte, n *NodeData, depth int, nodeRefcount, valueRefcount uint32) IterateNodesAction {
			touchedRefcounts.setNodeIfAbsent(n.Commitment, nodeRefcount)
			nodeRefcount = touchedRefcounts.incNode(n.Commitment)
			if nodeRefcount > 1 {
				return IterateSkipSubtree
			}
			if n.CommitsToExternalValue() {
				touchedRefcounts.setValueIfAbsent(n.Terminal, valueRefcount)
				touchedRefcounts.incValue(n.Terminal)
			}
			return IterateContinue
		})
		for hash, nodeRefcount := range touchedRefcounts.Nodes {
			tr.setNodeRefcount(hash, nodeRefcount)
		}
		for terminalData, valueRefcount := range touchedRefcounts.Values {
			t := &Tcommitment{Data: []byte(terminalData), IsValue: false}
			tr.setValueRefcount(t, valueRefcount)
		}
	}
	return nil
}
