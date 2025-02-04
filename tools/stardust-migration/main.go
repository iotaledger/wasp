// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"log"
	"os"

	"bytes"

	old_iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/stardust-migration/db"
	"github.com/iotaledger/stardust-migration/migrations"
	"github.com/iotaledger/stardust-migration/stateaccess/oldstate"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/state/indexedstore"
	old_isc "github.com/nnikolash/wasp-types-exported/packages/isc"
	old_kv "github.com/nnikolash/wasp-types-exported/packages/kv"
	old_collections "github.com/nnikolash/wasp-types-exported/packages/kv/collections"
	old_state "github.com/nnikolash/wasp-types-exported/packages/state"
	old_indexedstore "github.com/nnikolash/wasp-types-exported/packages/state/indexedstore"
	old_blocklog "github.com/nnikolash/wasp-types-exported/packages/vm/core/blocklog"
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
	if len(os.Args) < 4 {
		log.Fatalf("usage: %s <src-chain-db-dir> <dest-chain-db-dir> <new-chain-id>", os.Args[0])
	}

	srcChainDBDir := os.Args[1]
	destChainDBDir := os.Args[2]
	newChainIDStr := os.Args[3]

	lo.Must0(os.MkdirAll(destChainDBDir, 0755))

	entries := lo.Must(os.ReadDir(destChainDBDir))
	if len(entries) > 0 {
		log.Fatalf("destination directory is not empty: %v", destChainDBDir)
	}

	srcKVS := db.Connect(srcChainDBDir)
	srcStore := old_indexedstore.New(old_state.NewStoreWithUniqueWriteMutex(srcKVS))
	srcState := lo.Must(srcStore.LatestState())

	oldChainID := old_isc.ChainID(GetAnchorOutput(srcState).AliasID)
	newChainID := lo.Must(isc.ChainIDFromString(newChainIDStr))

	destKVS := db.Create(destChainDBDir)
	destStore := indexedstore.New(state.NewStoreWithUniqueWriteMutex(destKVS))
	destStateDraft := destStore.NewOriginStateDraft()

	migrations.MigrateRootContract(srcState, destStateDraft)
	migrations.MigrateAccountsContract(srcState, destStateDraft, oldChainID, newChainID)
	// migrations.MigrateBlocklogContract(srcState, destStateDraft)
	// migrations.MigrateGovernanceContract(srcState, destStateDraft)

	newBlock := destStore.Commit(destStateDraft)
	destStore.SetLatest(newBlock.TrieRoot())
	destKVS.Flush()
}

func GetAnchorOutput(chainState old_kv.KVStoreReader) *old_iotago.AliasOutput {
	contractState := oldstate.GetContactStateReader(chainState, old_blocklog.Contract.Hname())

	registry := old_collections.NewArrayReadOnly(contractState, old_blocklog.PrefixBlockRegistry)
	if registry.Len() == 0 {
		panic("Block registry is empty")
	}

	blockInfoBytes := registry.GetAt(registry.Len() - 1)

	var blockInfo old_blocklog.BlockInfo
	lo.Must0(blockInfo.Read(bytes.NewReader(blockInfoBytes)))

	return blockInfo.PreviousAliasOutput.GetAliasOutput()
}
