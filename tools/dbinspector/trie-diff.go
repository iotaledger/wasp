package main

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/wasp/packages/chaindb"
	"github.com/iotaledger/wasp/packages/trie"
)

type trieDiffStats struct {
	start     time.Time
	lastShown time.Time
	onlyOn1   int
	onlyOn2   int
	inCommon  int
}

func trieDiff(ctx context.Context, kvs kvstore.KVStore) {
	if blockIndex > blockIndex2 {
		blockIndex, blockIndex2 = blockIndex2, blockIndex
	}
	state1 := getState(kvs, blockIndex)
	tr1, err := trie.NewTrieReader(trie.NewHiveKVStoreAdapter(kvs, []byte{chaindb.PrefixTrie}), state1.TrieRoot())
	mustNoError(err)

	state2 := getState(kvs, blockIndex2)
	tr2, err := trie.NewTrieReader(trie.NewHiveKVStoreAdapter(kvs, []byte{chaindb.PrefixTrie}), state2.TrieRoot())
	mustNoError(err)

	diff := trieDiffStats{
		start:     time.Now(),
		lastShown: time.Now(),
	}

	type nodeData struct {
		*trie.NodeData
		key []byte
	}

	iterateTrie := func(tr *trie.TrieReader, ch chan *nodeData) {
		defer close(ch)
		tr.IterateNodes(func(nodeKey []byte, node *trie.NodeData, depth int) bool {
			if ctx.Err() != nil {
				return false
			}
			ch <- &nodeData{NodeData: node, key: nodeKey}
			return true
		})
	}
	ch1 := make(chan *nodeData, 100)
	ch2 := make(chan *nodeData, 100)
	go iterateTrie(tr1, ch1)
	go iterateTrie(tr2, ch2)

	// This is similar to the 'merge' function in mergeSort.
	// We iterate both tries in order, advancing the iterator of the smallest
	// node between the two.

	var current1 *nodeData
	var ok1 bool
	var current2 *nodeData
	var ok2 bool
	current1, ok1 = <-ch1
	current2, ok2 = <-ch2
	for ok1 && ok2 {
		if current1.Commitment == current2.Commitment {
			diff.inCommon++
			current1, ok1 = <-ch1
			current2, ok2 = <-ch2
		} else if bytes.Compare(current1.key, current2.key) < 0 {
			diff.onlyOn1++
			current1, ok1 = <-ch1
		} else {
			diff.onlyOn2++
			current2, ok2 = <-ch2
		}
		showDiff(false, &diff)
	}
	for ok1 {
		diff.onlyOn1++
		_, ok1 = <-ch1
		showDiff(false, &diff)
	}
	for ok2 {
		diff.onlyOn2++
		_, ok2 = <-ch2
		showDiff(false, &diff)
	}

	fmt.Print("\n--- Done ---\n")
	showDiff(true, &diff)
}

func showDiff(force bool, diff *trieDiffStats) {
	now := time.Now()
	if !force && now.Sub(diff.lastShown) < 1*time.Second {
		return
	}
	diff.lastShown = now

	n1 := diff.onlyOn1 + diff.inCommon
	n2 := diff.onlyOn2 + diff.inCommon

	clearScreen()
	fmt.Println()
	fmt.Printf("Diff between blocks #%d -> #%d\n", blockIndex, blockIndex2)
	fmt.Println()
	fmt.Printf("amount of nodes: %d -> %d (%+d)\n", n1, n2, n2-n1)
	fmt.Printf("in common: %d (%.2f%% of #%d) (%.2f%% of #%d)\n",
		diff.inCommon,
		percentf(diff.inCommon, n1),
		blockIndex,
		percentf(diff.inCommon, n2),
		blockIndex2,
	)
	fmt.Printf("only on #%d: %d (%.2f%%)\n",
		blockIndex,
		diff.onlyOn1,
		percentf(diff.onlyOn1, n1),
	)
	fmt.Printf("only on #%d: %d (%.2f%%)\n",
		blockIndex2,
		diff.onlyOn2,
		percentf(diff.onlyOn2, n2),
	)
	fmt.Println()
	elapsed := time.Since(diff.start)
	fmt.Printf("Elapsed: %s\n", elapsed)
	fmt.Printf("Speed: %d nodes/s\n", int(float64(n1+n2)/(elapsed.Seconds())))
}
