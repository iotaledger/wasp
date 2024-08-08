// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/iotaledger/hive.go/kvstore"
	hivedb "github.com/iotaledger/hive.go/kvstore/database"
	"github.com/iotaledger/hive.go/kvstore/rocksdb"
	"github.com/iotaledger/wasp/packages/database"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/subrealm"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/state/indexedstore"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
)

func main() {
	if len(os.Args) < 3 {
		log.Fatalf("usage: %s <src-chain-db-dir> <dest-chain-db-dir>", os.Args[0])
	}

	srcChainDBDir := os.Args[1]
	destChainDBDir := os.Args[2]

	must(os.MkdirAll(destChainDBDir, 0755))

	entries := must2(os.ReadDir(destChainDBDir))
	if len(entries) > 0 {
		log.Fatalf("destination directory is not empty: %v", destChainDBDir)
	}

	destKVS := createDB(destChainDBDir)
	destStore := indexedstore.New(state.NewStoreWithUniqueWriteMutex(destKVS))
	destStateDraft := destStore.NewOriginStateDraft()

	srcKVS := connectDB(srcChainDBDir)
	srcStore := indexedstore.New(state.NewStoreWithUniqueWriteMutex(srcKVS))
	srcState := must2(srcStore.LatestState())

	//migrateAccountsContractState(srcState, destStateDraft)
	//migrate<Other Contract>State(srcState, destStateDraft)

	migrateAccountsContractState2(srcState, destStateDraft)

	newBlock := destStore.Commit(destStateDraft)
	destStore.SetLatest(newBlock.TrieRoot())
	destKVS.Flush()
}

func createDB(dbDir string) kvstore.KVStore {
	// TODO: does this need any options?
	rocksDatabase := must2(rocksdb.CreateDB(dbDir))

	db := database.New(
		dbDir,
		rocksdb.New(rocksDatabase),
		hivedb.EngineRocksDB,
		true,
		func() bool { panic("should not be called") },
	)

	kvs := db.KVStore()

	return kvs
}

func connectDB(dbDir string) kvstore.KVStore {
	rocksDatabase, err := rocksdb.OpenDBReadOnly(dbDir,
		rocksdb.IncreaseParallelism(runtime.NumCPU()-1),
		rocksdb.Custom([]string{
			"periodic_compaction_seconds=43200",
			"level_compaction_dynamic_level_bytes=true",
			"keep_log_file_num=2",
			"max_log_file_size=50000000", // 50MB per log file
		}),
	)
	must(err)

	db := database.New(
		dbDir,
		rocksdb.New(rocksDatabase),
		hivedb.EngineRocksDB,
		true,
		func() bool { panic("should not be called") },
	)

	kvs := db.KVStore()

	return kvs
}

func must(err error) {
	if err != nil {
		panic(fmt.Sprintf("%+v", err))
	}
}

func must2[RetVal any](retVal RetVal, err error) RetVal {
	if err != nil {
		panic(fmt.Sprintf("%+v", err))
	}

	return retVal
}

type contactKeyInfo struct {
	KeyPrefix   string
	Description string
}

func getContactStateReader(chainState kv.KVStoreReader, contractHname isc.Hname) kv.KVStoreReader {
	return subrealm.NewReadOnly(chainState, kv.Key(contractHname.Bytes()))
}

func getContactState(chainState kv.KVStore, contractHname isc.Hname) kv.KVStore {
	return subrealm.New(chainState, kv.Key(contractHname.Bytes()))
}

func migrateContractState(chainState kv.KVStoreReader, contractName string, keys []contactKeyInfo, destChainState state.StateDraft) {
	contractStateReader := getContactStateReader(chainState, coreutil.CoreHname(contractName))

	log.Printf("Migrating contract state for '%v'...\n", coreutil.CoreHname(contractName))

	for _, keyInfo := range keys {
		log.Printf("Migrating '%v' ('%v')...\n", keyInfo.Description, keyInfo.KeyPrefix)

		count := 0
		contractStateReader.Iterate(kv.Key(keyInfo.KeyPrefix), func(key kv.Key, value []byte) bool {
			count++
			err := getMigrationHandler(keyInfo.KeyPrefix)(key, value, destChainState)
			if err != nil {
				panic(fmt.Sprintf("error migrating key '%v' ('%v''%v'): %+v", key.Hex(), keyInfo.Description, keyInfo.KeyPrefix, err))
			}
			return true
		})

		log.Printf("Migrated %v keys of '%v' ('%v')\n", count, keyInfo.Description, keyInfo.KeyPrefix)
	}
}

func migrateAccountsContractState(srcChainState kv.KVStoreReader, destChainState state.StateDraft) {
	keysToMigrate := []contactKeyInfo{
		{keyAllAccounts, "All accounts"},
		//{prefixBaseTokens + AccountID, "Base tokens by account"},
		{prefixBaseTokens + L2TotalsAccount, "L2 total base tokens"},
		//{PrefixNativeTokens + AccountID, "Native tokens by account"},
		{PrefixNativeTokens + L2TotalsAccount, "L2 total native tokens"},
		{PrefixNFTs, "NFTs per account"},
		{PrefixNFTsByCollection, "NFTs by collection"},
		{prefixNewlyMintedNFTs, "Newly minted NFTs"},
		{prefixMintIDMap, "Mint ID map"},
		{keyNFTOwner, "NFT owner"},
		{PrefixFoundries, "Foundries of accounts"},
		{noCollection, "No collection"},
		{keyNonce, "Nonce"},
		{keyNativeTokenOutputMap, "Native token output map"},
		{keyFoundryOutputRecords, "Foundry output records"},
		{keyNFTOutputRecords, "NFT output records"},
		{keyNewNativeTokens, "New native tokens"},
		{prefixUnprocessableRequests, "Unprocessable requests"},
		{prefixNewUnprocessableRequests, "New unprocessable requests"},
		// {VarChainOwnerID, "Chain owner ID"},
		// {VarChainOwnerIDDelegated, "Chain owner ID delegated"},
		// {VarMinBaseTokensOnCommonAccount, "Min base tokens on common account"},
		// {VarPayoutAgentID, "Payout agent ID"},
	}

	migrateContractState(srcChainState, accounts.Contract.Name, keysToMigrate, destChainState)
}

// func parseKey(key kv.Key) (isc.Hname, string) {
// 	hname := must2(isc.HnameFromBytes([]byte(key[:4])))
// 	postfix := string(key[4:])
// 	return hname, postfix
// }

func migrateAccountsContractState2(srcChainState kv.KVStoreReader, destChainState state.StateDraft) {
	srcContractState := getContactStateReader(srcChainState, coreutil.CoreHname(accounts.Contract.Name))
	destContractState := getContactState(destChainState, coreutil.CoreHname(accounts.Contract.Name))

	// Accounts
	log.Printf("Migrating accounts...\n")
	oldAgentIDToNewAgentID := map[kv.Key]kv.Key{}

	count := migrateEntitiesMapByName(srcContractState, destContractState, keyAllAccounts, "", p(func(oldAgentID kv.Key, srcVal bool) (kv.Key, bool) {
		newAgentID, newV := migrateAccount(oldAgentID, srcVal)
		oldAgentIDToNewAgentID[oldAgentID] = newAgentID

		return newAgentID, newV
	}))

	log.Printf("Migrated %v accounts\n", count)

	// All foundries
	log.Printf("Migrating list of all foundries...\n")

	count = migrateEntitiesMapByName(srcContractState, destContractState, keyFoundryOutputRecords, "", p(migrateFoundryOutput))

	log.Printf("Migrated %v foundries\n", count)

	// Founries per account
	// mapMame := PrefixFoundries + string(agentID.Bytes())
	log.Printf("Migrating foundries of accounts...\n")

	count = 0
	migrateFoundriesOfAccount := p(migrateFoundriesOfAccount)
	for oldAgentID, newAgentID := range oldAgentIDToNewAgentID {
		oldMapName := PrefixFoundries + string(oldAgentID)
		newMapName := PrefixFoundries + string(newAgentID)

		count += migrateEntitiesMapByName(srcContractState, destContractState, oldMapName, newMapName, migrateFoundriesOfAccount)
	}

	log.Printf("Migrated %v foundries of accounts\n", count)

	// Base token blances
	log.Printf("Migrating base token balances...\n")

	count = migrateEntitiesByPrefix(srcContractState, destContractState, prefixBaseTokens, p(migrateBaseTokenBalance))

	log.Printf("Migrated %v base token balances\n", count)

	// Native token balances
	// mapName := PrefixNativeTokens + string(accountKey)
	log.Printf("Migrating native token balances...\n")

	count = 0
	migrateNativeTokenBalance := p(migrateNativeTokenBalance)
	for oldAgentID, newAgentID := range oldAgentIDToNewAgentID {
		oldMapName := PrefixNativeTokens + string(oldAgentID)
		newMapName := PrefixNativeTokens + string(newAgentID)
		count += migrateEntitiesMapByName(srcContractState, destContractState, oldMapName, newMapName, migrateNativeTokenBalance)
	}

	log.Printf("Migrated %v native token balances\n", count)

	// Account to NFT
	// mapName := PrefixNFTs + string(agentID.Bytes())
	log.Printf("Migrating NFTs per account...\n")

	count = 0
	migrateNFT := p(migrateNFT)
	for oldAgentID, newAgentID := range oldAgentIDToNewAgentID {
		oldMapName := PrefixNFTs + string(oldAgentID)
		newMapName := PrefixNFTs + string(newAgentID)
		count += migrateEntitiesMapByName(srcContractState, destContractState, oldMapName, newMapName, migrateNFT)
	}

	log.Printf("Migrated %v NFTs per account\n", count)

	// NFT to Owner
	log.Printf("Migrating NFT owners...\n")

	count = migrateEntitiesMapByName(srcContractState, destContractState, keyNFTOwner, "", p(migrateNFTOwners))

	log.Printf("Migrated %v NFT owners\n", count)

	// NFTs by collection
	// mapName := PrefixNFTsByCollection + string(agentID.Bytes()) + string(collectionID.Bytes())
	log.Printf("Migrating NFTs by collection...\n")

	// NOTE: There is no easy way to retrieve list of referenced collections
	count = 0
	for oldAgentID, newAgentID := range oldAgentIDToNewAgentID {
		oldPrefix := PrefixNFTsByCollection + string(oldAgentID)
		count += migrateEntitiesByPrefix(srcContractState, destContractState, oldPrefix, func(oldKey kv.Key, srcVal bool) (destKey kv.Key, destVal bool) {
			return migrateNFTByCollection(oldKey, srcVal, oldAgentID, newAgentID)
		})
	}

	log.Printf("Migrated %v NFTs by collection\n", count)

	// Native token outputs
	log.Printf("Migrating native token outputs...\n")

	count = migrateEntitiesMapByName(srcContractState, destContractState, keyNativeTokenOutputMap, "", p(migrateNativeTokenOutput))

	log.Printf("Migrated %v native token outputs\n", count)
}

func migrateAsIs(srcKey kv.Key, srcVal []byte) (destKey kv.Key, destVal []byte) {
	return srcKey, srcVal
}

func migrateAccount(srcKey kv.Key, srcVal bool) (destKey kv.Key, destVal bool) {
	return srcKey, srcVal
}

func migrateFoundryOutput(srcKey kv.Key, srcVal foundryOutputRec) (destKey kv.Key, destVal string) {
	return srcKey, "dummy new value"
}

func migrateFoundriesOfAccount(srcKey kv.Key, srcVal bool) (destKey kv.Key, destVal bool) {
	return srcKey + "dummy new key", srcVal
}

func migrateBaseTokenBalance(srcKey kv.Key, srcVal []byte) (destKey kv.Key, destVal []byte) {
	return srcKey + "dummy new key", srcVal
}

func migrateNativeTokenBalance(srcKey kv.Key, srcVal []byte) (destKey kv.Key, destVal []byte) {
	return srcKey + "dummy new key", srcVal
}

func migrateNFT(srcKey kv.Key, srcVal bool) (destKey kv.Key, destVal bool) {
	return srcKey + "dummy new key", srcVal
}

func migrateNFTOwners(srcKey kv.Key, srcVal []byte) (destKey kv.Key, destVal []byte) {
	return srcKey + "dummy new key", append(srcVal, []byte("dummy new value")...)
}

func migrateNFTByCollection(oldKey kv.Key, srcVal bool, oldAgentID, newAgentID kv.Key) (destKey kv.Key, destVal bool) {
	oldMapName, oldMapElemKey := SplitMapKey(oldKey)

	oldPrefix := PrefixNFTsByCollection + string(oldAgentID)
	collectionID := oldMapName[len(oldPrefix):]

	newMapName := PrefixNFTsByCollection + string(newAgentID) + string(collectionID)

	newKey := newMapName

	if oldMapElemKey != "" {
		// If this record is map element - we form map element key.
		nftID := oldMapElemKey
		// TODO: migrate NFT ID
		newKey += "." + string(nftID)
	}

	return kv.Key(newKey), srcVal
}

func migrateNativeTokenOutput(srcKey kv.Key, srcVal nativeTokenOutputRec) (destKey kv.Key, destVal nativeTokenOutputRec) {
	return srcKey, srcVal
}
