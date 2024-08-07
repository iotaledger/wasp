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

	migrateAccountsContractState(srcState, destStateDraft)
	//migrate<Other Contract>State(srcState, destStateDraft)

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

func parseKey(key kv.Key) (isc.Hname, string) {
	hname := must2(isc.HnameFromBytes([]byte(key[:4])))
	postfix := string(key[4:])
	return hname, postfix
}
