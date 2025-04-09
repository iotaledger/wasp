package main

import (
	"context"
	"math"
	"os"
	"sync/atomic"

	"github.com/samber/lo"
	cmd "github.com/urfave/cli/v2"
	"golang.org/x/sync/errgroup"

	"github.com/iotaledger/wasp/tools/stardust-migration/blockindex"
	"github.com/iotaledger/wasp/tools/stardust-migration/stateaccess/oldstate"
	"github.com/iotaledger/wasp/tools/stardust-migration/utils/cli"
	"github.com/iotaledger/wasp/tools/stardust-migration/utils/db"

	old_kv "github.com/nnikolash/wasp-types-exported/packages/kv"
	old_collections "github.com/nnikolash/wasp-types-exported/packages/kv/collections"
	old_dict "github.com/nnikolash/wasp-types-exported/packages/kv/dict"
	old_state "github.com/nnikolash/wasp-types-exported/packages/state"
	old_indexedstore "github.com/nnikolash/wasp-types-exported/packages/state/indexedstore"
	old_accounts "github.com/nnikolash/wasp-types-exported/packages/vm/core/accounts"
	old_evm "github.com/nnikolash/wasp-types-exported/packages/vm/core/evm"
	old_evmimpl "github.com/nnikolash/wasp-types-exported/packages/vm/core/evm/evmimpl"
)

type StateContainsTargetCheckFunc func(state old_kv.KVStoreReader, onFound func(k old_kv.Key, v []byte) bool)

func search(name string, f StateContainsTargetCheckFunc) func(c *cmd.Context) error {
	return func(c *cmd.Context) error {
		chainDBDir := c.Args().Get(0)
		fromIndex := uint32(c.Uint64("from-block"))
		toIndex := uint32(c.Uint64("to-block"))
		findAll := c.Bool("all")
		threadsCount := uint32(c.Uint("parallel"))

		kvs := db.ConnectOld(chainDBDir)
		store := old_indexedstore.New(old_state.NewStoreWithUniqueWriteMutex(kvs))

		if toIndex == 0 {
			cli.Logf("Using latest new state index")
			toIndex = lo.Must(store.LatestBlockIndex())
		}

		cli.Logf("Searching for %v in blocks [%d ; %d]", name, fromIndex, toIndex)
		searchLinear(c.Context, name, store, fromIndex, toIndex, findAll, threadsCount, f)

		return nil
	}
}

func searchLinear(ctx context.Context, name string, store old_indexedstore.IndexedStore, fromIndex, toIndex uint32, findAll bool, threadsCount uint32, f StateContainsTargetCheckFunc) {
	indexer := blockindex.New(store)
	e := errgroup.Group{}

	found := atomic.Bool{}
	firstContainingBlockIndexes := lo.RepeatBy(int(threadsCount), func(i int) uint32 { return math.MaxUint32 })
	firstFoundKeys := make([]old_kv.Key, threadsCount)
	firstFoundValues := make([][]byte, threadsCount)
	lastCheckedBlockIndexes := make([]uint32, threadsCount)
	totalBlocks := toIndex - fromIndex + 1
	printProgress, done := cli.NewProgressPrinter("blocks", totalBlocks)

	cli.Logf("Starting %v search threads", threadsCount)

	for i := uint32(0); i < threadsCount; i++ {
		i := i

		e.Go(func() error {
			for blockIndex := fromIndex + i; blockIndex <= toIndex; blockIndex += threadsCount {
				block, _ := indexer.BlockByIndex(blockIndex)

				state := old_dict.Dict(block.Mutations().Sets)
				f(state, func(k old_kv.Key, v []byte) bool {
					found.Store(true)

					if firstFoundValues[i] == nil {
						firstContainingBlockIndexes[i] = blockIndex
						firstFoundKeys[i] = k
						firstFoundValues[i] = v
					}

					if findAll {
						cli.Logf("Found %v: block = %v, key = %x", name, blockIndex, []byte(k))
					}

					return findAll
				})

				lastCheckedBlockIndexes[i] = blockIndex

				if i == 0 && (blockIndex/threadsCount)%100 == 0 {
					printProgress(threadsCount * 100)
				}

				if found.Load() && !findAll {
					return nil
				}
				if ctx.Err() != nil {
					return nil
				}
			}

			return nil
		})
	}

	e.Wait()
	done()

	if ctx.Err() != nil {
		cli.Logf("Interrupted. Last checked block: %d", lo.Min(lastCheckedBlockIndexes))
		os.Exit(1)
	}

	_, earliestThreadIdx := lo.MinIndex(firstContainingBlockIndexes)
	if firstFoundValues[earliestThreadIdx] == nil {
		cli.Logf("No %v found in blocks [%d; %d]\n", name, fromIndex, toIndex)
		return
	}

	if !findAll {
		earliestBlockIndex := firstContainingBlockIndexes[earliestThreadIdx]
		earliestKey := firstFoundKeys[earliestThreadIdx]
		earliestValue := firstFoundValues[earliestThreadIdx]

		cli.Logf("Found %v FIRST occurrence:\nBlock index: %v\nKey = %x\nValue = %x", name, earliestBlockIndex, []byte(earliestKey), earliestValue)
	}
}

// Although works, it does not make much sense - the object could be created on one block and immediatelly deleted on next.
// So it's impossible to find it by binary search...
// func searchBinary(ctx context.Context, name string, store old_indexedstore.IndexedStore, fromIndex, toIndex uint32, f StateContainsTargetCheckFunc) error {
// 	index := blockindex.New(store)

// 	lastFoundKey, lastFoundValue := old_kv.Key(""), []byte{}
// 	cli.Logf("Searching range: [%v ; %v]", fromIndex, toIndex)

// 	for {
// 		middleIndex := (fromIndex + toIndex) / 2
// 		cli.Logf("Searching at index %v", middleIndex)

// 		_, trieRoot := index.BlockByIndex(middleIndex)
// 		state := lo.Must(store.StateByTrieRoot(trieRoot))

// 		found := false
// 		f(state, func(k old_kv.Key, v []byte) bool {
// 			found = true
// 			lastFoundKey = k
// 			lastFoundValue = v
// 			return false
// 		})

// 		if fromIndex == toIndex {
// 			break
// 		}
// 		if found {
// 			toIndex = middleIndex
// 			cli.Logf("Found %v in block %d - new range = [%d; %d]", name, middleIndex, fromIndex, toIndex)
// 		} else {
// 			fromIndex = middleIndex + 1
// 			cli.Logf("Not found %v in block %d - new range = [%d; %d]", name, middleIndex, fromIndex, toIndex)
// 		}
// 	}

// 	if lastFoundValue != nil {
// 		cli.Logf("Found %v in block %d:\nKey = %x\nValue = %x", name, fromIndex, lastFoundKey, lastFoundValue)
// 	} else {
// 		cli.Logf("No %v found in blocks [%d; %d]\n", name, fromIndex, toIndex)
// 	}

// 	return nil
// }

func searchISCMagicAllowance(chainState old_kv.KVStoreReader, onFound func(k old_kv.Key, v []byte) bool) {
	contractState := old_evm.ISCMagicSubrealmR(old_evm.ContractPartitionR(chainState))
	contractState.Iterate(old_evmimpl.PrefixAllowance, onFound)
}

func searchNFT(chainState old_kv.KVStoreReader, onFound func(k old_kv.Key, v []byte) bool) {
	contractState := oldstate.GetContactStateReader(chainState, old_accounts.Contract.Hname())
	nfts := old_collections.NewMapReadOnly(contractState, old_accounts.KeyNFTOutputRecords)
	nfts.Iterate(func(k, v []byte) bool { onFound(old_kv.Key(k), v); return false })
}
