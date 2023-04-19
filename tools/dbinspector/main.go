// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"runtime"

	"github.com/iotaledger/hive.go/kvstore"
	hivedb "github.com/iotaledger/hive.go/kvstore/database"
	"github.com/iotaledger/hive.go/kvstore/rocksdb"
	"github.com/iotaledger/wasp/packages/database"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/state/indexedstore"
)

type processFunc func(context.Context, kvstore.KVStore)

var blockIndex int64

func main() {
	flag.Int64Var(&blockIndex, "b", -1, "Block index")
	flag.Parse()

	if flag.NArg() != 2 {
		log.Fatalf("usage: %s [-b index] <command> <chain-db-dir>", os.Args[0])
	}
	args := flag.Args()
	var f processFunc
	switch args[0] {
	case "state-stats-per-hname":
		f = stateStatsPerHname
	case "trie-stats":
		f = trieStats
	default:
		log.Fatalf("unknown command: %s", args[0])
	}

	process(args[1], f)
}

func getState(kvs kvstore.KVStore) state.State {
	store := indexedstore.New(state.NewStore(kvs))
	if blockIndex < 0 {
		state, err := store.LatestState()
		mustNoError(err)
		return state
	}
	state, err := store.StateByIndex(uint32(blockIndex))
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

func mustNoError(err error) {
	if err != nil {
		panic(err)
	}
}
