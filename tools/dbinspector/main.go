// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	hivedb "github.com/iotaledger/hive.go/kvstore/database"
	"github.com/iotaledger/wasp/packages/database"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/state/indexedstore"
	"github.com/iotaledger/wasp/packages/vm/core/corecontracts"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatalf("usage: %s <chain-db-dir>", os.Args[0])
	}
	process(os.Args[1])
}

func process(dbDir string) {
	db, err := database.DatabaseWithDefaultSettings(dbDir, false, hivedb.EngineAuto, false)
	if err != nil {
		panic(err)
	}

	kvs := db.KVStore()

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{}, 1)

	go func() {
		defer close(done)

		store := indexedstore.New(state.NewStore(kvs))
		state, err := store.LatestState()
		if err != nil {
			panic(err)
		}
		dumpStateStats(ctx, state)
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

func dumpStateStats(ctx context.Context, state state.State) {
	totalSize := 0

	var seenHnames []isc.Hname
	hnameUsedSpace := make(map[isc.Hname]int)
	hnameCount := make(map[isc.Hname]int)

	show := func() {
		fmt.Printf("\n\n Total DB size: %d\n\n", totalSize)
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
