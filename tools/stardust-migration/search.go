package main

import (
	"os"

	"github.com/samber/lo"
	cmd "github.com/urfave/cli/v2"

	"github.com/iotaledger/wasp/tools/stardust-migration/utils/cli"
	"github.com/iotaledger/wasp/tools/stardust-migration/utils/db"

	old_kv "github.com/nnikolash/wasp-types-exported/packages/kv"
	old_state "github.com/nnikolash/wasp-types-exported/packages/state"
	old_indexedstore "github.com/nnikolash/wasp-types-exported/packages/state/indexedstore"
	old_trie "github.com/nnikolash/wasp-types-exported/packages/trie"
	old_evm "github.com/nnikolash/wasp-types-exported/packages/vm/core/evm"
	old_evmimpl "github.com/nnikolash/wasp-types-exported/packages/vm/core/evm/evmimpl"
)

func search(name string, checkStateContainsTarget func(state old_kv.KVStoreReader, onFound func(k old_kv.Key, v []byte) bool)) func(c *cmd.Context) error {
	return func(c *cmd.Context) error {
		chainDBDir := c.Args().Get(0)
		fromIndex := uint32(c.Uint64("from-block"))
		toIndex := uint32(c.Uint64("to-block"))
		findAll := c.Bool("all")

		kvs := db.ConnectOld(chainDBDir)
		store := old_indexedstore.New(old_state.NewStoreWithUniqueWriteMutex(kvs))

		if toIndex == 0 {
			cli.Logf("Using latest new state index")
			toIndex = lo.Must(store.LatestBlockIndex())
		}

		cli.Logf("Searching for %v in blocks %d-%d\n", name, fromIndex, toIndex)
		foundAtLeastOne := false

		forEachBlock(store, fromIndex, toIndex, func(blockIndex uint32, blockHash old_trie.Hash, block old_state.Block) bool {
			if c.Err() != nil {
				cli.ClearStatusBar()
				cli.Logf("Interrupted. Last checked block: %d", blockIndex-1)
				os.Exit(1)
			}

			checkStateContainsTarget(dictKvFromMuts(block.Mutations()), func(k old_kv.Key, v []byte) bool {
				cli.Logf("Found %v:\n\tBlock index: %v\n\tKey = %x\n\t%Value = %x", name, k, v)
				if !findAll {
					os.Exit(0)
				}

				foundAtLeastOne = true
				return true
			})

			return true
		})

		if !foundAtLeastOne {
			cli.Logf("No %v found in blocks %d-%d\n", name, fromIndex, toIndex)
		}

		return nil
	}
}

func searchISCMagicAllowance(state old_kv.KVStoreReader, onFound func(k old_kv.Key, v []byte) bool) {
	oldMagicState := old_evm.ISCMagicSubrealmR(old_evm.ContractPartitionR(state))
	oldMagicState.Iterate(old_evmimpl.PrefixAllowance, onFound)
}
