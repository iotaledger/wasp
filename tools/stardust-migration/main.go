// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"log"
	"math/big"
	"os"

	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/state/indexedstore"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/core/migrations/allmigrations"
	old_isc "github.com/nnikolash/wasp-types-exported/packages/isc"
	old_kv "github.com/nnikolash/wasp-types-exported/packages/kv"
	old_collections "github.com/nnikolash/wasp-types-exported/packages/kv/collections"
	old_state "github.com/nnikolash/wasp-types-exported/packages/state"
	old_indexedstore "github.com/nnikolash/wasp-types-exported/packages/state/indexedstore"
	old_accounts "github.com/nnikolash/wasp-types-exported/packages/vm/core/accounts"
	old_blocklog "github.com/nnikolash/wasp-types-exported/packages/vm/core/blocklog"
	old_governance "github.com/nnikolash/wasp-types-exported/packages/vm/core/governance"
	"github.com/samber/lo"
)

// NOTE: Every entity type should be explicitly included in migration
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

	migrateAccountsContract(srcState, destStateDraft)
	// migrateBlocklogContract(srcState, destStateDraft)
	// migrateGovernanceContract(srcState, destStateDraft)

	newBlock := destStore.Commit(destStateDraft)
	destStore.SetLatest(newBlock.TrieRoot())
	destKVS.Flush()
}

var (
	// TODO: what is the correct value?
	oldBaseTokenDecimals uint32 = 6

	schema = allmigrations.SchemaVersionIotaRebased
)

func migrateAccountsContract(srcChainState old_kv.KVStoreReader, destChainState state.StateDraft) {
	log.Print("Migrating accounts contract...\n")

	srcState := getContactStateReader(srcChainState, old_accounts.Contract.Hname())
	destState := getContactState(destChainState, accounts.Contract.Hname())

	oldChainID := old_isc.ChainID(GetAnchorOutput(srcChainState).AliasID)
	newChainID := OldChainIDToNewChainID(oldChainID)

	oldAgentIDToNewAgentID := map[old_isc.AgentID]isc.AgentID{}

	migrateAccountsList(srcState, destState, oldChainID, newChainID, &oldAgentIDToNewAgentID)
	migrateBaseTokenBalances(srcState, destState, oldChainID, oldAgentIDToNewAgentID)
	migrateNativeTokenBalances(srcState, destState, oldChainID, oldAgentIDToNewAgentID)
	// migrateFoundriesOutputs(srcState, destState)
	// migrateFoundriesPerAccount(srcState, destState, oldAgentIDToNewAgentID)
	// migrateAccountToNFT(srcState, destState, oldAgentIDToNewAgentID)
	// migrateNFTtoOwner(srcState, destState)
	// migrateNFTsByCollection(srcState, destState, oldAgentIDToNewAgentID)
	// migrateNativeTokenOutputs(srcState, destState)
	// migrateNativeTokenBalanceTotal(srcState, destState)
	// migrateAllMintedNfts(srcState, destState)

	log.Print("Migrated accounts contract\n")
}

func migrateBlocklogContract(srcChainState old_kv.KVStoreReader, destChainState state.StateDraft) {
	log.Print("Migrating blocklog contract\n")

	// Unprocessable Requests (blocklog contract)
	// No need to migrate. Just print a warning if there are any
	log.Printf("Listing Unprocessable Requests...\n")

	blocklogContractStateSrc := getContactStateReader(srcChainState, old_blocklog.Contract.Hname())
	count := 0
	old_collections.NewMapReadOnly(blocklogContractStateSrc, old_blocklog.PrefixUnprocessableRequests).Iterate(func(srcKey, srcBytes []byte) bool {
		reqID := lo.Must(Deserialize[isc.RequestID](srcKey))
		log.Printf("Warning: unprocessable request found %v", reqID.String())
		count++
		return true
	})

	log.Printf("Listing Unprocessable Requests completed (found %v entities)\n", count)

	log.Print("Migrated blocklog contract\n")
}

func migrateGovernanceContract(srcChainState old_kv.KVStoreReader, destChainState state.StateDraft) {
	srcState := getContactStateReader(srcChainState, old_governance.Contract.Hname())
	destState := getContactState(destChainState, governance.Contract.Hname())

	log.Print("Migrating governance contract\n")

	// Chain Owner
	log.Printf("Migrating chain owner...\n")
	migrateEntityState(srcState, destState, old_governance.VarChainOwnerID, migrateAsIs(""))
	log.Printf("Migrated chain owner\n")

	// Chain Owner delegated
	log.Printf("Migrating chain owner delegated...\n")
	migrateEntityState(srcState, destState, old_governance.VarChainOwnerIDDelegated, migrateAsIs(""))
	log.Printf("Migrated chain owner delegated\n")

	// Payout agent
	log.Printf("Migrating Payout agent...\n")
	migrateEntityState(srcState, destState, old_governance.VarPayoutAgentID, migrateAsIs(""))
	log.Printf("Migrated Payout agent\n")

	// Min Base Tokens On Common Account
	log.Printf("Migrating Min Base Tokens On Common Account...\n")
	migrateEntityState(srcState, destState, old_governance.VarMinBaseTokensOnCommonAccount, migrateAsIs(""))
	log.Printf("Migrated Min Base Tokens On Common Account\n")

	log.Print("Migrated governance contract\n")
}

func migrateAccountsList(srcState old_kv.KVStoreReader, destState kv.KVStore, oldChID old_isc.ChainID, newChID isc.ChainID, oldAgentIDToNewAgentID *map[old_isc.AgentID]isc.AgentID) {
	log.Printf("Migrating accounts list...\n")

	migrateAccountAndSaveNewAgentID := p(func(oldAccountKey old_kv.Key, srcVal bool) (kv.Key, bool) {
		oldAgentID := lo.Must(old_accounts.AgentIDFromKey(oldAccountKey, oldChID))
		newAgentID := OldAgentIDtoNewAgentID(oldAgentID)

		// Pointer is not needed here, but its here just to emphasize that it is output argument.
		(*oldAgentIDToNewAgentID)[oldAgentID] = newAgentID

		return accounts.AccountKey(newAgentID, newChID), srcVal
	})

	count := migrateEntitiesMapByName(
		srcState, destState,
		old_accounts.KeyAllAccounts, accounts.KeyAllAccounts,
		migrateAccountAndSaveNewAgentID,
	)

	log.Printf("Migrated list of %v accounts\n", count)
}

func migrateBaseTokenBalances(srcState old_kv.KVStoreReader, destState kv.KVStore, oldChainID old_isc.ChainID, oldAgentIDToNewAgentID map[old_isc.AgentID]isc.AgentID) {
	log.Printf("Migrating base token balances...\n")

	// NOTE: Simply iterating by prefix is unsafe - the prefix might be a subprefix of another prefix

	for oldAgentID, newAgentID := range oldAgentIDToNewAgentID {
		oldKey := old_accounts.BaseTokensKey(old_accounts.AccountKey(oldAgentID, oldChainID))
		oldValBytes := srcState.Get(oldKey)
		// TODO: Could it be missing? How that would work?
		oldAmount := DecodeOldTokens(oldValBytes)

		newAccountKey := accounts.AccountKey(newAgentID, OldChainIDToNewChainID(oldChainID))
		newAmount := OldTokensCountToNewCoinValue(oldAmount)

		coinBalances := collections.NewMap(destState, accounts.AccountCoinBalancesKey(newAccountKey))
		coinBalances.SetAt(coin.BaseTokenType.Bytes(), codec.Encode(newAmount))
	}

	log.Printf("Migrated %v base token balances\n", len(oldAgentIDToNewAgentID))
}

func migrateNativeTokenBalances(srcState old_kv.KVStoreReader, destState kv.KVStore, oldChainID old_isc.ChainID, oldAgentIDToNewAgentID map[old_isc.AgentID]isc.AgentID) {
	log.Printf("Migrating native token balances...\n")

	var count uint32

	// NOTE: Simply iterating by prefix is unsafe - the prefix might be a subprefix of another prefix

	for oldAgentID, newAgentID := range oldAgentIDToNewAgentID {
		oldTokensMap := old_accounts.NativeTokensMapR(srcState, old_accounts.AccountKey(oldAgentID, oldChainID))

		oldTokensMap.Iterate(func(oldNativeTokenIDBytes []byte, oldTokenAmountBytes []byte) bool {
			oldNativeTokenID := old_isc.MustNativeTokenIDFromBytes(oldNativeTokenIDBytes)
			// TODO: Does base and native token have same encoding?...
			oldAmount := DecodeOldTokens(oldTokenAmountBytes)

			newCoinType := OldNativeTokemIDtoNewCoinType(oldNativeTokenID)
			newAccountKey := accounts.AccountKey(newAgentID, OldChainIDToNewChainID(oldChainID))
			newAmount := OldTokensCountToNewCoinValue(oldAmount)

			coinBalances := collections.NewMap(destState, accounts.AccountCoinBalancesKey(newAccountKey))
			coinBalances.SetAt(newCoinType.Bytes(), codec.Encode(newAmount))

			count++

			return true
		})
	}

	log.Printf("Migrated %v native token balances\n", count)
}

func migrateFoundriesOutputs(srcState old_kv.KVStoreReader, destState kv.KVStore) {
	log.Printf("Migrating list of foundry outputs...\n")

	migrateEntry := func(srcKey old_kv.Key, srcVal old_accounts.FoundryOutputRec) (destKey kv.Key, destVal string) {
		return kv.Key(srcKey), "dummy new value"
	}

	count := migrateEntitiesMapByName(srcState, destState, old_accounts.KeyFoundryOutputRecords, "", p(migrateEntry))

	log.Printf("Migrated %v foundry outputs\n", count)
}

func migrateFoundriesPerAccount(srcState old_kv.KVStoreReader, destState kv.KVStore, oldAgentIDToNewAgentID map[old_isc.AgentID]isc.AgentID) {
	log.Printf("Migrating foundries of accounts...\n")

	var count uint32

	migrateFoundriesOfAccount := p(func(srcKey old_kv.Key, srcVal bool) (destKey kv.Key, destVal bool) {
		return kv.Key(srcKey) + "dummy new key", srcVal
	})

	for oldAgentID, newAgentID := range oldAgentIDToNewAgentID {
		// mapMame := PrefixFoundries + string(agentID.Bytes())
		oldMapName := old_accounts.PrefixFoundries + string(oldAgentID.Bytes())
		_ = newAgentID
		newMapName := "" //accounts.PrefixFoundries + string(newAgentID.Bytes())

		count += migrateEntitiesMapByName(srcState, destState, oldMapName, newMapName, migrateFoundriesOfAccount)
	}

	log.Printf("Migrated %v foundries of accounts\n", count)
}

func migrateAccountToNFT(srcState old_kv.KVStoreReader, destState kv.KVStore, oldAgentIDToNewAgentID map[old_isc.AgentID]isc.AgentID) {
	log.Printf("Migrating NFTs per account...\n")

	var count uint32
	migrateEntry := p(func(srcKey old_kv.Key, srcVal bool) (destKey kv.Key, destVal bool) {
		return kv.Key(srcKey) + "dummy new key", srcVal
	})

	for oldAgentID, newAgentID := range oldAgentIDToNewAgentID {
		// mapName := PrefixNFTs + string(agentID.Bytes())
		oldMapName := old_accounts.PrefixNFTs + string(oldAgentID.Bytes())
		_ = newAgentID
		newMapName := "" // accounts.PrefixNFTs + string(newAgentID.Bytes())

		count += migrateEntitiesMapByName(srcState, destState, oldMapName, newMapName, migrateEntry)
	}

	log.Printf("Migrated %v NFTs per account\n", count)
}

func migrateNFTtoOwner(srcState old_kv.KVStoreReader, destState kv.KVStore) {
	log.Printf("Migrating NFT owners...\n")

	migrateEntry := func(srcKey old_kv.Key, srcVal []byte) (destKey kv.Key, destVal []byte) {
		return kv.Key(srcKey) + "dummy new key", append(srcVal, []byte("dummy new value")...)
	}

	count := migrateEntitiesMapByName(srcState, destState, old_accounts.KeyNFTOwner, "", p(migrateEntry))
	log.Printf("Migrated %v NFT owners\n", count)
}

func migrateNFTsByCollection(srcState old_kv.KVStoreReader, destState kv.KVStore, oldAgentIDToNewAgentID map[old_isc.AgentID]isc.AgentID) {
	log.Printf("Migrating NFTs by collection...\n")

	var count uint32

	for oldAgentID, newAgentID := range oldAgentIDToNewAgentID {
		// mapName := PrefixNFTsByCollection + string(agentID.Bytes()) + string(collectionID.Bytes())
		// NOTE: There is no easy way to retrieve list of referenced collections
		oldPrefix := old_accounts.PrefixNFTsByCollection + string(oldAgentID.Bytes())

		count += migrateEntitiesByPrefix(srcState, destState, oldPrefix, func(oldKey old_kv.Key, srcVal bool) (destKey kv.Key, destVal bool) {
			return migrateNFTsByCollectionEntry(oldKey, srcVal, oldAgentID, newAgentID)
		})
	}

	log.Printf("Migrated %v NFTs by collection\n", count)
}

func migrateNFTsByCollectionEntry(oldKey old_kv.Key, srcVal bool, oldAgentID old_isc.AgentID, newAgentID isc.AgentID) (destKey kv.Key, destVal bool) {
	oldMapName, oldMapElemKey := SplitMapKey(oldKey)

	oldPrefix := old_accounts.PrefixNFTsByCollection + string(oldAgentID.Bytes())
	collectionIDBytes := oldMapName[len(oldPrefix):]
	_ = collectionIDBytes

	newMapName := "" // accounts.PrefixNFTsByCollection + string(newAgentID.Bytes()) + string(collectionIDBytes)

	newKey := newMapName

	if oldMapElemKey != "" {
		// If this record is map element - we form map element key.
		nftID := oldMapElemKey
		// TODO: migrate NFT ID
		newKey += "." + string(nftID)
	}

	return kv.Key(newKey), srcVal
}

func migrateNativeTokenOutputs(srcState old_kv.KVStoreReader, destState kv.KVStore) {
	log.Printf("Migrating native token outputs...\n")

	migrateEntry := func(srcKey old_kv.Key, srcVal old_accounts.NativeTokenOutputRec) (destKey kv.Key, destVal old_accounts.NativeTokenOutputRec) {
		return kv.Key(srcKey), srcVal
	}

	count := migrateEntitiesMapByName(srcState, destState, old_accounts.KeyNativeTokenOutputMap, "", p(migrateEntry))

	log.Printf("Migrated %v native token outputs\n", count)
}

func migrateNativeTokenBalanceTotal(srcState old_kv.KVStoreReader, destState kv.KVStore) {
	log.Printf("Migrating native token total balance...\n")

	migrateEntry := func(srcKey old_kv.Key, srcVal *big.Int) (destKey kv.Key, destVal []byte) {
		// TODO: new amount format (if not big.Int)
		return kv.Key(srcKey), []byte{0}
	}

	count := migrateEntitiesMapByName(srcState, destState, old_accounts.PrefixNativeTokens+accounts.L2TotalsAccount, "", p(migrateEntry))

	log.Printf("Migrated %v native token total balance\n", count)
}

func migrateAllMintedNfts(srcState old_kv.KVStoreReader, destState kv.KVStore) {
	// prefixMintIDMap stores a map of <internal NFTID> => <NFTID>
	log.Printf("Migrating All minted NFTs...\n")

	migrateEntry := func(srcKey old_kv.Key, srcVal []byte) (destKey kv.Key, destVal []byte) {
		return kv.Key(srcKey), []byte{0}
	}

	count := migrateEntitiesMapByName(srcState, destState, old_accounts.PrefixMintIDMap, "", p(migrateEntry))

	log.Printf("Migrated %v All minted NFTs\n", count)
}
