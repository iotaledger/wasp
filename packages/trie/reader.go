package trie

import (
	"encoding/hex"
	"fmt"
	"io"
	"maps"
	"strings"
)

// Reader direct read-only access to trie
type Reader struct {
	nodeStore *nodeStore
	root      Hash
}

func NewReader(store KVReader, root Hash) *Reader {
	nodeStore := openNodeStore(store)
	return &Reader{
		nodeStore: nodeStore,
		root:      root,
	}
}

func (tr *Reader) Root() Hash {
	return tr.root
}

func (tr *Reader) VerifyRoot() error {
	_, ok := tr.nodeStore.FetchNodeData(tr.root)
	if !ok {
		return fmt.Errorf("trie root not found: %s", tr.root)
	}
	return nil
}

// DebugDump prints the structure of the tree to stdout, for debugging purposes.
func (tr *Reader) DebugDump(w io.Writer, nodeCounts map[Hash]uint32, valueCounts map[string]uint32) {
	tr.IterateNodes(func(path []byte, n *NodeData, depth int) IterateNodesAction {
		nodeCount := nodeCounts[n.Commitment]
		nodeCount++
		nodeCounts[n.Commitment] = nodeCount

		key := "[]"
		if len(path) > 0 {
			key = fmt.Sprintf("[%d]", path[len(path)-1])
		}
		indent := strings.Repeat(" ", depth*4)
		fmt.Fprintf(w, "%s %v %s (seen: %d)\n", indent, key, n, nodeCount)

		if nodeCount > 1 {
			return IterateSkipSubtree
		}
		if n.Terminal != nil && !n.Terminal.IsValue {
			valueCount := valueCounts[string(n.Terminal.Data)]
			valueCount++
			valueCounts[string(n.Terminal.Data)] = valueCount

			fmt.Fprintf(
				w,
				"%s     [v: %x -> %q] (seen: %d)\n",
				indent,
				n.Terminal.Data,
				ellipsis(tr.nodeStore.valueStore.Get(n.Terminal.Bytes()), 20),
				valueCount,
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

// DebugDump prints the structure of the whole DB to stdout, for debugging
// purposes. It also verifies the refcounts, and panics if there is a mismatch.
func DebugDump(store KVStore, roots []Hash, w io.Writer) {
	nodeCounts := make(map[Hash]uint32)
	valueCounts := make(map[string]uint32)

	fmt.Fprintf(w, "[trie store]\n")
	for _, root := range roots {
		tr := NewReader(store, root)
		tr.DebugDump(w, nodeCounts, valueCounts)
	}

	enabled, refcounts := NewRefcounts(store)
	if !enabled {
		fmt.Fprint(w, "[node refcounts disabled]\n")
		return
	}
	nodeCounts2, valueCounts2 := refcounts.DebugDump(w)
	if !maps.Equal(nodeCounts, nodeCounts2) {
		showDiff(w, nodeCounts, nodeCounts2, func(h Hash) string { return h.String() })
		panic("inconsistency: node counts do not match")
	}
	if !maps.Equal(valueCounts, valueCounts2) {
		showDiff(w, valueCounts, valueCounts2, func(s string) string { return hex.EncodeToString([]byte(s)) })
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
