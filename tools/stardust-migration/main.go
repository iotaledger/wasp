// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"log"
	"os"

	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/state/indexedstore"
	old_state "github.com/nnikolash/wasp-types-exported/packages/state"
	old_indexedstore "github.com/nnikolash/wasp-types-exported/packages/state/indexedstore"
	"github.com/samber/lo"
)

// NOTE: Every record type should be explicitly included in migration
// NOTE: All migration is node at once or just abandoned. There is no option to continue.
// TODO: Do we start from block 0 or N+1 where N last old block?
// TODO: Do we prune old block? Are we going to do migration from origin? If not, have we pruned blocks with old schemas?
// TODO: What to do with foundry prefixes?
// TODO: From where to get new chain ID?
// TODO: Need to migrate ALL trie roots to support tracing.
// TODO: New state draft might be huge, but it is stored in memory - might be an issue.

func main() {
	if len(os.Args) < 3 {
		log.Fatalf("usage: %s <src-chain-db-dir> <dest-chain-db-dir>", os.Args[0])
	}

	srcChainDBDir := os.Args[1]
	destChainDBDir := os.Args[2]

	lo.Must0(os.MkdirAll(destChainDBDir, 0755))

	entries := lo.Must(os.ReadDir(destChainDBDir))
	if len(entries) > 0 {
		log.Fatalf("destination directory is not empty: %v", destChainDBDir)
	}

	destKVS := createDB(destChainDBDir)
	destStore := indexedstore.New(state.NewStoreWithUniqueWriteMutex(destKVS))
	destStateDraft := destStore.NewOriginStateDraft()

	srcKVS := connectDB(srcChainDBDir)
	srcStore := old_indexedstore.New(old_state.NewStoreWithUniqueWriteMutex(srcKVS))
	srcState := lo.Must(srcStore.LatestState())

	migrateRootContract(srcState, destStateDraft)
	migrateAccountsContract(srcState, destStateDraft)
	// migrateBlocklogContract(srcState, destStateDraft)
	// migrateGovernanceContract(srcState, destStateDraft)

	newBlock := destStore.Commit(destStateDraft)
	destStore.SetLatest(newBlock.TrieRoot())
	destKVS.Flush()
}
