// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"

	"fortio.org/safecast"

	hivedb "github.com/iotaledger/hive.go/db"
	"github.com/iotaledger/wasp/v2/packages/database"
	"github.com/iotaledger/wasp/v2/packages/kvstore"
	"github.com/iotaledger/wasp/v2/packages/kvstore/rocksdb"
	"github.com/iotaledger/wasp/v2/packages/state"
	"github.com/iotaledger/wasp/v2/packages/state/indexedstore"
	"github.com/iotaledger/wasp/v2/packages/state/statetest"
)

type processFunc func(context.Context, kvstore.KVStore)

var (
	blockIndex  int64
	blockIndex2 int64
)

func main() {
	flag.Int64Var(&blockIndex, "b", -1, "Block index")
	flag.Int64Var(&blockIndex2, "B", -1, "Block index 2")
	flag.Parse()

	if flag.NArg() != 2 {
		log.Fatalf("usage: %s [-b index] <command> <chain-db-dir>", os.Args[0])
	}
	args := flag.Args()
	var f processFunc
	switch args[0] {
	// This stats function is for a Stardust ISC DB only!
	case "state-stats-per-hname-with-keys":
		f = stateStatsPerHnameWithKeys
	case "state-stats-per-hname":
		f = stateStatsPerHname
	case "trie-stats":
		f = trieStats
	case "trie-diff":
		f = trieDiff
	default:
		log.Fatalf("unknown command: %s", args[0])
	}

	process(args[1], f)
}

func getState(kvs kvstore.KVStore, index int64) state.State {
	store := indexedstore.New(statetest.NewStoreWithUniqueWriteMutex(kvs))
	if index < 0 {
		state, err := store.LatestState()
		mustNoError(err)
		return state
	}
	indexUint32, err := safecast.Convert[uint32](index)
	mustNoError(err)
	state, err := store.StateByIndex(indexUint32)
	mustNoError(err)
	return state
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

func percent(a, n int) int {
	return int(percentf(a, n))
}

func percentf(a, n int) float64 {
	return (100.0 * float64(a)) / float64(n)
}

func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

func mustNoError(err error) {
	if err != nil {
		panic(err)
	}
}
