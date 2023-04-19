// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"os/signal"
	"runtime"
	"time"

	"github.com/iotaledger/hive.go/kvstore"
	hivedb "github.com/iotaledger/hive.go/kvstore/database"
	"github.com/iotaledger/hive.go/kvstore/rocksdb"
	"github.com/iotaledger/wasp/packages/chaindb"
	"github.com/iotaledger/wasp/packages/database"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/state/indexedstore"
	"github.com/iotaledger/wasp/packages/trie"
	"github.com/iotaledger/wasp/packages/vm/core/corecontracts"
)

type processFunc func(context.Context, kvstore.KVStore)

func main() {
	if len(os.Args) != 3 {
		log.Fatalf("usage: %s <command> <chain-db-dir>", os.Args[0])
	}
	var f processFunc
	switch os.Args[1] {
	case "latest-block-state-per-hname":
		f = latestBlockStatePerHname
	case "latest-block-trie":
		f = latestBlockTrie
	default:
		log.Fatalf("unknown command: %s", os.Args[1])
	}

	process(os.Args[2], f)
}

func process(dbDir string, f processFunc) {
	rocksDatabase, err := rocksdb.OpenDBReadOnly(dbDir,
		rocksdb.IncreaseParallelism(runtime.NumCPU()-1),
		rocksdb.Custom([]string{
			"periodic_compaction_seconds=43200",
			"level_compaction_dynamic_level_bytes=true",
			"keep_log_file_num=2",
			"max_log_file_size=50000000", // 50MB per log file
		}),
	)
	mustNoError(err)

	db := database.New(
		dbDir,
		rocksdb.New(rocksDatabase),
		hivedb.EngineRocksDB,
		true,
		func() bool { panic("should not be called") },
	)
	kvs := db.KVStore()

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{}, 1)

	go func() {
		defer close(done)

		start := time.Now()
		defer func() {
			fmt.Printf("\nTook: %s\n", time.Since(start))
		}()

		f(ctx, kvs)
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	select {
	case <-c:
		cancel()
		<-done
	case <-done:
		cancel()
	}
}

func latestBlockTrie(ctx context.Context, kvs kvstore.KVStore) {
	store := indexedstore.New(state.NewStore(kvs))
	state, err := store.LatestState()
	mustNoError(err)
	tr, err := trie.NewTrieReader(trie.NewHiveKVStoreAdapter(kvs, []byte{chaindb.PrefixTrie}), state.TrieRoot())
	mustNoError(err)

	n := 0
	size := 0
	var childCount [trie.NumChildren + 1]int
	terminal := 0
	notTerminal := 0
	terminalIsValue := 0
	depthSum := 0

	percent := func(a, n int) int {
		return int(math.Round((100.0 * float64(a)) / float64(n)))
	}

	show := func() {
		fmt.Print("\033[H\033[2J") // clear screen
		fmt.Printf("\nTotal trie nodes: %d\n", n)
		fmt.Println()
		fmt.Printf(" not terminal: %d (%2d%%)\n", notTerminal, percent(notTerminal, n))
		fmt.Printf("  is terminal: %d (%2d%%)\n", terminal, percent(terminal, n))
		fmt.Printf("     terminal is value: %d (%2d%% of terminal nodes)\n", terminalIsValue, percent(terminalIsValue, terminal))
		fmt.Println()
		for i := 0; i <= trie.NumChildren; i++ {
			fmt.Printf(" with %2d children: %9d (%2d%%)\n", i, childCount[i], percent(childCount[i], n))
		}
		fmt.Println()
		fmt.Printf("Total size: %d bytes\n", size)
		fmt.Printf("Avg node size: %d bytes\n", size/n)
		fmt.Printf("Avg node depth: %.2f\n", float32(depthSum)/float32(n))
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

func latestBlockStatePerHname(ctx context.Context, kvs kvstore.KVStore) {
	store := indexedstore.New(state.NewStore(kvs))
	state, err := store.LatestState()
	mustNoError(err)

	totalSize := 0

	var seenHnames []isc.Hname
	hnameUsedSpace := make(map[isc.Hname]int)
	hnameCount := make(map[isc.Hname]int)

	show := func() {
		fmt.Printf("State index: %d\n", state.BlockIndex())
		fmt.Printf("Total state size: %d bytes\n\n", totalSize)
		for _, hn := range seenHnames {
			hns := hn.String()
			if corecontracts.All[hn] != nil {
				hns = corecontracts.All[hn].Name
			}
			fmt.Printf("%s: %d key-value pairs -- size: %d bytes\n", hns, hnameCount[hn], hnameUsedSpace[hn])
		}
	}

	n := 0
	state.IterateSorted("", func(k kv.Key, v []byte) bool {
		if ctx.Err() != nil {
			fmt.Println(ctx.Err())
			return false
		}
		if len(k) < 4 {
			fmt.Printf("len(k) < 4: %x\n", k)
			return true
		}
		usedSpace := len(k) + len(v)
		totalSize += usedSpace
		hn, err := isc.HnameFromBytes([]byte(k[:4]))
		if err == nil {
			if hnameCount[hn] == 0 {
				seenHnames = append(seenHnames, hn)
			}
			hnameUsedSpace[hn] += usedSpace
			hnameCount[hn]++
		}
		n++
		if n%10000 == 0 {
			show()
		}
		return true
	})
	show()
}

func mustNoError(err error) {
	if err != nil {
		panic(err)
	}
}
