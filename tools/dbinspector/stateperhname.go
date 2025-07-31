package main

import (
	"context"
	"fmt"

	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/kv"
	"github.com/iotaledger/wasp/v2/packages/kvstore"
	"github.com/iotaledger/wasp/v2/packages/vm/core/corecontracts"
)

func stateStatsPerHname(ctx context.Context, kvs kvstore.KVStore) {
	state := getState(kvs, blockIndex)

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
