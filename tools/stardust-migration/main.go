// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"log"
	"math/big"
	"os"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/state/indexedstore"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
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
// TODO: New state draft might be huge, but it is stored in memory - might be an issue.
// TODO: Do we start from block 0 or N+1 where N last old block?
// TODO: Need to migrate ALL trie roots to support tracing.

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

	//migrateAccountsContractState(srcState, destStateDraft)
	//migrate<Other Contract>State(srcState, destStateDraft)

	migrateAccountsContractState(srcState, destStateDraft)
	migrateOtherContractStates(srcState, destStateDraft)

	newBlock := destStore.Commit(destStateDraft)
	destStore.SetLatest(newBlock.TrieRoot())
	destKVS.Flush()
}

func migrateAccountsContractState(srcChainState old_kv.KVStoreReader, destChainState state.StateDraft) {
	srcContractState := getContactStateReader(srcChainState, old_accounts.Contract.Hname())
	destContractState := getContactState(destChainState, accounts.Contract.Hname())

	oldChainID := old_isc.ChainID(GetAnchorOutput(srcChainState).AliasID)
	newChainID := isc.ChainID(oldChainID)

	log.Print("Migrating accounts contract state...\n")

	// Accounts
	log.Printf("Migrating accounts...\n")
	oldAgentIDToNewAgentID := map[old_isc.AgentID]isc.AgentID{}

	count := migrateEntitiesMapByName(srcContractState, destContractState, old_accounts.KeyAllAccounts, accounts.KeyAllAccounts, p(func(oldAccountKey old_kv.Key, srcVal bool) (kv.Key, bool) {
		oldAgentID := lo.Must(old_accounts.AgentIDFromKey(oldAccountKey, oldChainID))
		newAgentID, newV := migrateAccount(oldAgentID, srcVal)
		oldAgentIDToNewAgentID[oldAgentID] = newAgentID

		return accounts.AccountKey(newAgentID, newChainID), newV
	}))

	log.Printf("Migrated %v accounts\n", count)

	// // All foundries
	// log.Printf("Migrating list of all foundries...\n")
	// count = migrateEntitiesMapByName(srcContractState, destContractState, accounts.KeyFoundryOutputRecords, "", p(migrateFoundryOutput))
	// log.Printf("Migrated %v foundries\n", count)

	// // Founries per account
	// // mapMame := PrefixFoundries + string(agentID.Bytes())
	// log.Printf("Migrating foundries of accounts...\n")

	// count = 0
	// migrateFoundriesOfAccount := p(migrateFoundriesOfAccount)
	// for oldAgentID, newAgentID := range oldAgentIDToNewAgentID {
	// 	oldMapName := accounts.PrefixFoundries + string(oldAgentID.Bytes())
	// 	newMapName := accounts.PrefixFoundries + string(newAgentID.Bytes())

	// 	count += migrateEntitiesMapByName(srcContractState, destContractState, oldMapName, newMapName, migrateFoundriesOfAccount)
	// }
	// log.Printf("Migrated %v foundries of accounts\n", count)

	// // Base token blances
	// log.Printf("Migrating base token balances...\n")

	// count = migrateEntitiesByPrefix(srcContractState, destContractState, accounts.PrefixBaseTokens, p(migrateBaseTokenBalance))

	// log.Printf("Migrated %v base token balances\n", count)

	// // Native token balances
	// // mapName := PrefixNativeTokens + string(accountKey)
	// log.Printf("Migrating native token balances...\n")

	// count = 0
	// migrateNativeTokenBalance := p(migrateNativeTokenBalance)
	// for oldAgentID, newAgentID := range oldAgentIDToNewAgentID {
	// 	oldMapName := accounts.PrefixNativeTokens + string(accounts.AccountKey(oldAgentID, oldChainID))
	// 	newMapName := accounts.PrefixNativeTokens + string(accounts.AccountKey(newAgentID, oldChainID))
	// 	count += migrateEntitiesMapByName(srcContractState, destContractState, oldMapName, newMapName, migrateNativeTokenBalance)
	// }

	// log.Printf("Migrated %v native token balances\n", count)

	// // Account to NFT
	// // mapName := PrefixNFTs + string(agentID.Bytes())
	// log.Printf("Migrating NFTs per account...\n")

	// count = 0
	// migrateNFT := p(migrateNFT)
	// for oldAgentID, newAgentID := range oldAgentIDToNewAgentID {
	// 	oldMapName := accounts.PrefixNFTs + string(oldAgentID.Bytes())
	// 	newMapName := accounts.PrefixNFTs + string(newAgentID.Bytes())
	// 	count += migrateEntitiesMapByName(srcContractState, destContractState, oldMapName, newMapName, migrateNFT)
	// }

	// log.Printf("Migrated %v NFTs per account\n", count)

	// // NFT to Owner
	// log.Printf("Migrating NFT owners...\n")

	// count = migrateEntitiesMapByName(srcContractState, destContractState, accounts.KeyNFTOwner, "", p(migrateNFTOwners))

	// log.Printf("Migrated %v NFT owners\n", count)

	// // NFTs by collection
	// // mapName := PrefixNFTsByCollection + string(agentID.Bytes()) + string(collectionID.Bytes())
	// log.Printf("Migrating NFTs by collection...\n")

	// // NOTE: There is no easy way to retrieve list of referenced collections
	// count = 0
	// for oldAgentID, newAgentID := range oldAgentIDToNewAgentID {
	// 	oldPrefix := accounts.PrefixNFTsByCollection + string(oldAgentID.Bytes())
	// 	count += migrateEntitiesByPrefix(srcContractState, destContractState, oldPrefix, func(oldKey kv.Key, srcVal bool) (destKey kv.Key, destVal bool) {
	// 		return migrateNFTByCollection(oldKey, srcVal, oldAgentID, newAgentID)
	// 	})
	// }

	// log.Printf("Migrated %v NFTs by collection\n", count)

	// // Native token outputs
	// log.Printf("Migrating native token outputs...\n")

	// count = migrateEntitiesMapByName(srcContractState, destContractState, accounts.KeyNativeTokenOutputMap, "", p(migrateNativeTokenOutput))

	// log.Printf("Migrated %v native token outputs\n", count)

	// // Native token total balance
	// log.Printf("Migrating native token total balance...\n")

	// count = migrateEntitiesMapByName(srcContractState, destContractState, accounts.PrefixNativeTokens+accounts.L2TotalsAccount, "", p(migrateNativeTokenBalanceTotal))

	// log.Printf("Migrated %v native token total balance\n", count)

	// // All minted NFTs
	// // prefixMintIDMap stores a map of <internal NFTID> => <NFTID>
	// log.Printf("Migrating All minted NFTs...\n")

	// count = migrateEntitiesMapByName(srcContractState, destContractState, accounts.PrefixMintIDMap, "", p(migrateAllMintedNfts))

	// log.Printf("Migrated %v All minted NFTs\n", count)

	log.Print("Migrated accounts contract state\n")
}

func migrateOtherContractStates(srcChainState old_kv.KVStoreReader, destChainState state.StateDraft) {

	//srcContractState := getContactStateReader(srcChainState, coreutil.CoreHname(blocklog.Contract.Name))
	// destContractState := getContactState(destChainState, coreutil.CoreHname(accounts.Contract.Name))

	governanceContractStateSrc := getContactStateReader(srcChainState, old_governance.Contract.Hname())
	governanceContractStateDest := getContactState(destChainState, governance.Contract.Hname())

	log.Print("Migrating other contracts states\n")

	// Unprocessable Requests (blocklog contract)
	// No need to migrate. Just print a warning if there are any
	log.Printf("Listing Unprocessable Requests...\n")

	blocklogContractStateSrc := getContactStateReader(srcChainState, old_blocklog.Contract.Hname())
	count := 0
	old_collections.NewMapReadOnly(blocklogContractStateSrc, old_blocklog.PrefixUnprocessableRequests).Iterate(func(srcKey, srcBytes []byte) bool {
		reqID := lo.Must(DeserializeEntity[isc.RequestID](srcKey))
		log.Printf("Warning: unprocessable request found %v", reqID.String())
		count++
		return true
	})
	log.Printf("Listing Unprocessable Requests completed (found %v entities)\n", count)

	// Chain Owner
	log.Printf("Migrating chain owner...\n")
	migrateEntityState(governanceContractStateSrc, governanceContractStateDest, old_governance.VarChainOwnerID, migrateAsIs)
	log.Printf("Migrated chain owner\n")

	// Chain Owner delegated
	log.Printf("Migrating chain owner delegated...\n")
	migrateEntityState(governanceContractStateSrc, governanceContractStateDest, old_governance.VarChainOwnerIDDelegated, migrateAsIs)
	log.Printf("Migrated chain owner delegated\n")

	// Payout agent
	log.Printf("Migrating Payout agent...\n")
	migrateEntityState(governanceContractStateSrc, governanceContractStateDest, old_governance.VarPayoutAgentID, migrateAsIs)
	log.Printf("Migrated Payout agent\n")

	// Min Base Tokens On Common Account
	log.Printf("Migrating Min Base Tokens On Common Account...\n")
	migrateEntityState(governanceContractStateSrc, governanceContractStateDest, old_governance.VarMinBaseTokensOnCommonAccount, migrateAsIs)
	log.Printf("Migrated Min Base Tokens On Common Account\n")

	log.Print("Migrated other contracts states\n")
}

func migrateAsIs(srcKey old_kv.Key, srcVal []byte) (destKey kv.Key, destVal []byte) {
	return kv.Key(srcKey), srcVal
}

func migrateAccount(oldAgentID old_isc.AgentID, srcVal bool) (newAgentID isc.AgentID, destVal bool) {
	//return isc.AgentID(oldAgentID), srcVal
	panic("not implemented")
}

func migrateFoundryOutput(srcKey old_kv.Key, srcVal old_accounts.FoundryOutputRec) (destKey kv.Key, destVal string) {
	return kv.Key(srcKey), "dummy new value"
}

func migrateFoundriesOfAccount(srcKey old_kv.Key, srcVal bool) (destKey kv.Key, destVal bool) {
	return kv.Key(srcKey) + "dummy new key", srcVal
}

func migrateBaseTokenBalance(srcKey old_kv.Key, srcVal []byte) (destKey kv.Key, destVal []byte) {
	return kv.Key(srcKey) + "dummy new key", srcVal
}

func migrateNativeTokenBalance(srcKey old_kv.Key, srcVal []byte) (destKey kv.Key, destVal []byte) {
	return kv.Key(srcKey) + "dummy new key", srcVal
}

func migrateNFT(srcKey old_kv.Key, srcVal bool) (destKey kv.Key, destVal bool) {
	return kv.Key(srcKey) + "dummy new key", srcVal
}

func migrateNFTOwners(srcKey old_kv.Key, srcVal []byte) (destKey kv.Key, destVal []byte) {
	return kv.Key(srcKey) + "dummy new key", append(srcVal, []byte("dummy new value")...)
}

func migrateNFTByCollection(oldKey old_kv.Key, srcVal bool, oldAgentID, newAgentID isc.AgentID) (destKey kv.Key, destVal bool) {
	oldMapName, oldMapElemKey := SplitMapKey(oldKey)

	oldPrefix := old_accounts.PrefixNFTsByCollection + string(oldAgentID.Bytes())
	collectionIDBytes := oldMapName[len(oldPrefix):]

	newMapName := old_accounts.PrefixNFTsByCollection + string(newAgentID.Bytes()) + string(collectionIDBytes)

	newKey := newMapName

	if oldMapElemKey != "" {
		// If this record is map element - we form map element key.
		nftID := oldMapElemKey
		// TODO: migrate NFT ID
		newKey += "." + string(nftID)
	}

	return kv.Key(newKey), srcVal
}

func migrateNativeTokenOutput(srcKey old_kv.Key, srcVal old_accounts.NativeTokenOutputRec) (destKey kv.Key, destVal old_accounts.NativeTokenOutputRec) {
	return kv.Key(srcKey), srcVal
}

func migrateNativeTokenBalanceTotal(srcKey old_kv.Key, srcVal *big.Int) (destKey kv.Key, destVal []byte) {
	// TODO: new amount format (if not big.Int)
	return kv.Key(srcKey), []byte{0}
}

func migrateAllMintedNfts(srcKey old_kv.Key, srcVal []byte) (destKey kv.Key, destVal []byte) {
	return kv.Key(srcKey), []byte{0}
}
