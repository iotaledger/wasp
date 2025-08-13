package main

import (
	"context"
	"fmt"
	"time"

	"github.com/iotaledger/wasp/v2/packages/chaindb"
	"github.com/iotaledger/wasp/v2/packages/kvstore"
	"github.com/iotaledger/wasp/v2/packages/trie"
)

type trieStatsData struct {
	n               int
	size            int
	childCount      [trie.NumChildren + 1]int
	terminal        int
	notTerminal     int
	terminalIsValue int
	depthSum        int

	start     time.Time
	lastShown time.Time
}

func trieStats(ctx context.Context, kvs kvstore.KVStore) {
	state := getState(kvs, blockIndex)
	blockIndex = int64(state.BlockIndex())
	tr := trie.NewReader(trie.NewHiveKVStoreAdapter(kvs, []byte{chaindb.PrefixTrie}), state.TrieRoot())

	data := trieStatsData{
		start:     time.Now(),
		lastShown: time.Now(),
	}

	type nodeData struct {
		*trie.NodeData
		depth int
	}
	nodesCh := make(chan nodeData, 100)

	go func() {
		defer close(nodesCh)
		tr.IterateNodes(func(nodeKey []byte, node *trie.NodeData, depth int) trie.IterateNodesAction {
			if ctx.Err() != nil {
				fmt.Println(ctx.Err())
				return trie.IterateStop
			}
			nodesCh <- nodeData{NodeData: node, depth: depth}
			return trie.IterateContinue
		})
	}()

	last := time.Now()
	for node := range nodesCh {
		data.n++

		data.size += len(node.Bytes()) + trie.HashSizeBytes
		data.childCount[node.ChildrenCount()]++
		if node.Terminal == nil {
			data.notTerminal++
		} else {
			data.terminal++
			if node.Terminal.IsValue {
				data.terminalIsValue++
			}
		}
		data.depthSum += node.depth

		now := time.Now()
		if now.Sub(last) > 1*time.Second {
			clearScreen()
			showTrieStats(&data)
			last = now
		}
	}
	showTrieStats(&data)
}

func showTrieStats(data *trieStatsData) {
	fmt.Println()
	fmt.Printf("Block index: %d\n", blockIndex)
	fmt.Println()
	fmt.Printf("Total trie nodes: %d\n", data.n)
	fmt.Println()
	fmt.Printf("non-terminal: %9d (%2d%%)\n", data.notTerminal, percent(data.notTerminal, data.n))
	fmt.Printf("    terminal: %9d (%2d%%)\n", data.terminal, percent(data.terminal, data.n))
	fmt.Println()
	fmt.Printf("     value stored in node: %9d (%2d%% of terminal nodes)\n", data.terminalIsValue, percent(data.terminalIsValue, data.terminal))
	fmt.Printf("value stored outside node: %9d (%2d%% of terminal nodes)\n", data.terminal-data.terminalIsValue, percent(data.terminal-data.terminalIsValue, data.terminal))
	fmt.Println()
	for i := 0; i <= trie.NumChildren; i++ {
		fmt.Printf("with %2d children: %9d (%2d%%)\n", i, data.childCount[i], percent(data.childCount[i], data.n))
	}
	fmt.Println()
	fmt.Printf("Total trie size: %d bytes\n", data.size)
	fmt.Printf("Avg node size: %d bytes\n", data.size/data.n)
	fmt.Printf("Avg node depth: %.2f\n", float32(data.depthSum)/float32(data.n))
	fmt.Println()
	elapsed := time.Since(data.start)
	fmt.Printf("Elapsed: %s\n", elapsed)
	fmt.Printf("Speed: %d nodes/s\n", int(float64(data.n)/(elapsed.Seconds())))
}
