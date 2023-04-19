package main

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"time"

	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/wasp/packages/chaindb"
	"github.com/iotaledger/wasp/packages/trie"
)

func trieStats(ctx context.Context, kvs kvstore.KVStore) {
	state := getState(kvs)
	tr, err := trie.NewTrieReader(trie.NewHiveKVStoreAdapter(kvs, []byte{chaindb.PrefixTrie}), state.TrieRoot())
	mustNoError(err)

	n := 0
	size := 0
	var childCount [trie.NumChildren + 1]int
	terminal := 0
	notTerminal := 0
	terminalIsValue := 0
	depthSum := 0

	start := time.Now()

	percent := func(a, n int) int {
		return int(math.Round((100.0 * float64(a)) / float64(n)))
	}

	show := func() {
		fmt.Print("\033[H\033[2J") // clear screen
		fmt.Println()
		fmt.Printf("Block index: %d\n", state.BlockIndex())
		fmt.Println()
		fmt.Printf("Total trie nodes: %d\n", n)
		fmt.Println()
		fmt.Printf("non-terminal: %9d (%2d%%)\n", notTerminal, percent(notTerminal, n))
		fmt.Printf("    terminal: %9d (%2d%%)\n", terminal, percent(terminal, n))
		fmt.Println()
		fmt.Printf("     value stored in node: %9d (%2d%% of terminal nodes)\n", terminalIsValue, percent(terminalIsValue, terminal))
		fmt.Printf("value stored outside node: %9d (%2d%% of terminal nodes)\n", terminal-terminalIsValue, percent(terminal-terminalIsValue, terminal))
		fmt.Println()
		for i := 0; i <= trie.NumChildren; i++ {
			fmt.Printf("with %2d children: %9d (%2d%%)\n", i, childCount[i], percent(childCount[i], n))
		}
		fmt.Println()
		fmt.Printf("Total trie size: %d bytes\n", size)
		fmt.Printf("Avg node size: %d bytes\n", size/n)
		fmt.Printf("Avg node depth: %.2f\n", float32(depthSum)/float32(n))
		fmt.Println()
		elapsed := time.Since(start)
		fmt.Printf("Elapsed: %s\n", elapsed)
		fmt.Printf("Speed: %d nodes/s\n", int(float64(n)/(elapsed.Seconds())))
	}

	type nodeData struct {
		*trie.NodeData
		depth int
	}
	nodesCh := make(chan nodeData, 100)

	go func() {
		defer close(nodesCh)
		tr.IterateNodes(func(nodeKey []byte, node *trie.NodeData, depth int) bool {
			if ctx.Err() != nil {
				fmt.Println(ctx.Err())
				return false
			}
			nodesCh <- nodeData{NodeData: node, depth: depth}
			return true
		})
	}()

	last := time.Now()
	for node := range nodesCh {
		n++

		var buf bytes.Buffer
		err := node.Write(&buf)
		mustNoError(err)
		size += len(buf.Bytes()) + trie.HashSizeBytes
		childCount[node.ChildrenCount()]++
		if node.Terminal == nil {
			notTerminal++
		} else {
			terminal++
			if node.Terminal.IsValue {
				terminalIsValue++
			}
		}
		depthSum += node.depth

		now := time.Now()
		if now.Sub(last) > 1*time.Second {
			show()
			last = now
		}
	}
	show()
}
