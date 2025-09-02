package trie

import (
	"encoding/hex"
	"fmt"
	"io"
	"maps"
	"strings"
)

// TrieR provides read-only access to the trie
type TrieR struct {
	store KVReader
}

// TrieRFromRoot provides read-only access to the trie from a given trie root
type TrieRFromRoot struct {
	R    *TrieR
	Root Hash
}

// TrieRW provides read-write access to the trie
type TrieRW struct {
	*TrieR
	store KVStore
}

func NewTrieR(store KVReader) *TrieR {
	return &TrieR{store: store}
}

func NewTrieRFromRoot(store KVReader, root Hash) *TrieRFromRoot {
	return &TrieRFromRoot{
		R:    NewTrieR(makeCachedKVReader(store)),
		Root: root,
	}
}

func NewTrieRW(store KVStore) *TrieRW {
	return &TrieRW{
		TrieR: NewTrieR(store),
		store: store,
	}
}

// InitRoot initializes an empty trie store by committing a trie root
func (tr *TrieRW) InitRoot(refcountsEnabled bool) (Hash, error) {
	err := tr.UpdateRefcountsFlag(refcountsEnabled)
	if err != nil {
		return Hash{}, err
	}

	rootNodeData := newNodeData()
	n := newDraftNode(rootNodeData, nil)

	tr.commitNode(n)
	tr.initRefcounts(n)
	return n.nodeData.Commitment, nil
}

// VerifyRoot fetches the root node, and returns error if it is not found.
func (tr *TrieRFromRoot) VerifyRoot() error {
	_, ok := tr.R.fetchNodeData(tr.Root)
	if !ok {
		return fmt.Errorf("trie root not found: %s", tr.Root)
	}
	return nil
}

// DebugDump prints the structure of the trie to w, for debugging purposes.
func (tr *TrieRFromRoot) DebugDump(w io.Writer, refcounts *Refcounts) {
	tr.IterateNodes(func(path []byte, n *NodeData, depth int) IterateNodesAction {
		nodeRefcount := refcounts.incNode(n.Commitment)
		key := "[]"
		if len(path) > 0 {
			key = fmt.Sprintf("[%d]", path[len(path)-1])
		}
		indent := strings.Repeat(" ", depth*4)
		fmt.Fprintf(w, "%s %v %s (seen: %d)\n", indent, key, n, nodeRefcount)
		if nodeRefcount > 1 {
			return IterateSkipSubtree
		}
		if n.CommitsToExternalValue() {
			valueRefcount := refcounts.incValue(n.Terminal)
			fmt.Fprintf(
				w,
				"%s     [v: %x -> %q] (seen: %d)\n",
				indent,
				n.Terminal.Data,
				ellipsis(tr.R.fetchValueOfTerminal(n.Terminal), 20),
				valueRefcount,
			)
		}
		return IterateContinue
	})
}

func ellipsis(b []byte, maxLen int) string {
	if len(b) <= maxLen {
		return string(b)
	}
	if maxLen < 3 {
		maxLen = 3
	}
	return string(b[0:maxLen-3]) + "..."
}

// DebugDump prints the structure of the whole DB to w, for debugging
// purposes. It also verifies the refcounts, and panics if there is a mismatch.
func (tr *TrieR) DebugDump(roots []Hash, w io.Writer) {
	refcounts := NewRefcounts()

	fmt.Fprintf(w, "[trie store]\n")
	for _, root := range roots {
		NewTrieRFromRoot(tr.store, root).DebugDump(w, refcounts)
	}

	if !tr.IsRefcountsEnabled() {
		fmt.Fprint(w, "[node refcounts disabled]\n")
		return
	}
	refcounts2 := tr.DebugDumpRefcounts(w)
	if !maps.Equal(refcounts.Nodes, refcounts2.Nodes) {
		showDiff(w, refcounts.Nodes, refcounts2.Nodes, func(h Hash) string { return h.String() })
		panic("inconsistency: node counts do not match")
	}
	if !maps.Equal(refcounts.Values, refcounts2.Values) {
		showDiff(w, refcounts.Values, refcounts2.Values, func(s string) string { return hex.EncodeToString([]byte(s)) })
		panic("inconsistency: value counts do not match")
	}
}

func showDiff[T comparable](w io.Writer, a, b map[T]uint32, toString func(T) string) {
	fmt.Fprint(w, "[counts diff]\n")
	for k, v := range a {
		if vb, ok := b[k]; !ok || v != vb {
			fmt.Fprintf(w, "  <- %s: %d\n", toString(k), v)
		}
	}
	for k, v := range b {
		if va, ok := a[k]; !ok || v != va {
			fmt.Fprintf(w, "  -> %s: %d\n", toString(k), v)
		}
	}
}
