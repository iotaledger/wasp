package main

import (
	"context"
	"fmt"
	"time"

	"github.com/iotaledger/wasp/v2/packages/chaindb"
	"github.com/iotaledger/wasp/v2/packages/kvstore"
	"github.com/iotaledger/wasp/v2/packages/trie"
)

func trieDiff(ctx context.Context, kvs kvstore.KVStore) {
	if blockIndex > blockIndex2 {
		blockIndex, blockIndex2 = blockIndex2, blockIndex
	}
	state1 := getState(kvs, blockIndex)
	state2 := getState(kvs, blockIndex2)

	start := time.Now()

	onlyOn1, onlyOn2 := trie.NewTrieR(trie.NewHiveKVStoreAdapter(kvs, []byte{chaindb.PrefixTrie})).
		Diff(state1.TrieRoot(), state2.TrieRoot())

	fmt.Printf("Diff between blocks #%d -> #%d\n", blockIndex, blockIndex2)
	fmt.Printf("only on #%d: %d\n", blockIndex, len(onlyOn1))
	fmt.Printf("only on #%d: %d\n", blockIndex2, len(onlyOn2))
	elapsed := time.Since(start)
	fmt.Printf("Elapsed: %s\n", elapsed)
}
