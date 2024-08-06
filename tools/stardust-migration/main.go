// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"slices"

	"github.com/iotaledger/hive.go/kvstore"
	hivedb "github.com/iotaledger/hive.go/kvstore/database"
	"github.com/iotaledger/hive.go/kvstore/rocksdb"
	"github.com/iotaledger/wasp/packages/database"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/subrealm"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/state/indexedstore"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/corecontracts"
	"github.com/samber/lo"
)

func main() {
	if len(os.Args) < 3 {
		log.Fatalf("usage: %s <src-chain-db-dir> <dest-chain-db-dir>", os.Args[0])
	}

	srcChainDBDir := os.Args[1]
	destChainDBDir := os.Args[2]

	_ = srcChainDBDir

	must(os.MkdirAll(destChainDBDir, 0755))

	entries := must2(os.ReadDir(destChainDBDir))
	if len(entries) > 0 {
		log.Fatalf("destination directory is not empty: %v", destChainDBDir)
	}

	// srcKVS := connectDB(srcChainDBDir)
	// srcStore := indexedstore.New(state.NewStoreWithUniqueWriteMutex(srcKVS))
	// srcState, err := srcStore.LatestState()
	// must(err)

	destKVS := createDB(destChainDBDir)

	destStore := indexedstore.New(state.NewStoreWithUniqueWriteMutex(destKVS))

	stateDraft := destStore.NewOriginStateDraft()

	stateDraft.Set(kv.Key("test-key"), []byte("test-value"))
	newBlock := destStore.Commit(stateDraft)
	destStore.SetLatest(newBlock.TrieRoot())
	destKVS.Flush()

	destState := must2(destStore.LatestState())

	log.Printf("Has: %v", destState.Has(kv.Key("test-key")))
	log.Printf("Get: %v", destState.Get(kv.Key("test-key")))

	// showAnchorStats(srcState)
	// //showBlocklogContractStats(state)
	// showAccountsContractStats(srcState)
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

func getContactState(chainState kv.KVStoreReader, contractHname isc.Hname) kv.KVStoreReader {
	return subrealm.NewReadOnly(chainState, kv.Key(contractHname.Bytes()))
}

func showContractKeysStats(chainState kv.KVStoreReader, contractHname isc.Hname, keys []contactKeyInfo) {
	contractState := getContactState(chainState, contractHname)

	keys = slices.Clone(keys)

	for i := range keys {
		keyInfo := &keys[i]

		if keyInfo.Description == "" {
			keyInfo.Description = keyInfo.KeyPrefix
		} else {
			keyInfo.Description = keyInfo.Description + " (" + keyInfo.KeyPrefix + ")"
		}
	}

	searchResults := make(map[string]bool, len(keys))

	for _, keyInfo := range keys {
		log.Printf("Searching for %v...\n", keyInfo.Description)

		found := false
		contractState.IterateKeys(kv.Key(keyInfo.KeyPrefix), func(key kv.Key) bool {
			found = true
			return false
		})

		if found {
			log.Printf("%v: found\n", keyInfo.Description)
		} else {
			log.Printf("%v: no entries\n", keyInfo.Description)
		}

		searchResults[keyInfo.Description] = found
	}

	log.Printf("Search results:\n")
	for _, keyInfo := range keys {
		foundStr := lo.Ternary(searchResults[keyInfo.Description], "FOUND", "NOT FOUND")
		log.Printf("    %-25v\t%v\n", keyInfo.Description, foundStr)
	}

	for _, keyInfo := range keys {
		log.Printf("Counting %v...\n", keyInfo.Description)

		entriesCount := 0

		contractState.IterateKeys(kv.Key(keyInfo.KeyPrefix), func(key kv.Key) bool {
			entriesCount++

			if entriesCount%1000 == 0 {
				log.Printf("%v: %v entries\n", keyInfo.Description, entriesCount)
			}

			return true
		})

		log.Printf("Total %v: %v entries\n", keyInfo.Description, entriesCount)
	}
}

func showAccountsContractStats(chainState kv.KVStoreReader) {
	showContractKeysStats(chainState, accounts.Contract.Hname(), []contactKeyInfo{
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
	})
}

// func showAnchorStats(chainState kv.KVStoreReader) {
// 	contractState := getContactState(chainState, blocklog.Contract.Hname())

// 	if !contractState.Has(kv.Key(PrefixBlockRegistry)) {
// 		panic("Block registry not found in contract state")
// 	}

// 	registry := collections.NewArrayReadOnly(contractState, PrefixBlockRegistry)
// 	if registry.Len() == 0 {
// 		panic("Block registry is empty")
// 	}

// 	blockInfoBytes := registry.GetAt(registry.Len() - 1)

// 	var blockInfo blocklog.BlockInfo
// 	err := blockInfo.Read(bytes.NewReader(blockInfoBytes))
// 	must(err)

// 	log.Printf("Block index: %d\n", blockInfo.BlockIndex())
// 	log.Printf("FoundryCounter: %v\n", blockInfo.PreviousAliasOutput.GetAliasOutput().FoundryCounter)

// 	lastBlockWithNonZeroFoundryCounter := -1
// 	for i := int64(registry.Len()) - 1; i >= 0; i-- {
// 		blockInfoBytes := registry.GetAt(uint32(i))

// 		var blockInfo blocklog.BlockInfo
// 		err := blockInfo.Read(bytes.NewReader(blockInfoBytes))
// 		if err != nil {
// 			if errors.Is(err, io.EOF) {
// 				log.Printf("Reached on the blocklog: blockIndex = %d\n", i+1)
// 				break
// 			}

// 			must(err)
// 		}

// 		if blockInfo.PreviousAliasOutput.GetAliasOutput().FoundryCounter != 0 {
// 			lastBlockWithNonZeroFoundryCounter = int(i)
// 			break
// 		}
// 	}

// 	if lastBlockWithNonZeroFoundryCounter == -1 {
// 		log.Printf("No blocks with non-zero FoundryCounter found")
// 	} else {
// 		log.Printf("Last block with non-zero FoundryCounter: %d", lastBlockWithNonZeroFoundryCounter)
// 	}
// }

// func showBlocklogContractStats(chainState kv.KVStoreReader) {
// 	showContractKeysStats(chainState, blocklog.Contract.Hname(), []contactKeyInfo{
// 		{PrefixBlockRegistry, "Block registry"},
// 		{prefixRequestLookupIndex, "Request lookup index"},
// 		{prefixRequestReceipts, "Request receipts"},
// 		{prefixRequestEvents, "Request events"},
// 		{prefixUnprocessableRequests, "Unprocessable requests"},
// 		{prefixNewUnprocessableRequests, "New unprocessable requests"},
// 	})
// }

func iterateChainState(chainState kv.KVStoreReader) {
	maxIterations := 10

	//s.Iterate(kv.Key(contractHname.Bytes()), func(key kv.Key, value []byte) bool {
	chainState.IterateSorted("", func(key kv.Key, value []byte) bool {
		hn, postfix := parseKey(key)

		hns := hn.String()
		hnType := "(unknown)"
		if corecontracts.All[hn] != nil {
			hns = corecontracts.All[hn].Name
			hnType = "(core contract)"
			//contractState := subrealm.NewReadOnly(s, kv.Key(hn))
		}

		log.Printf("%v %v: %v\n", hns, hnType, postfix)

		maxIterations--

		return maxIterations > 0
	})
}

func parseKey(key kv.Key) (isc.Hname, string) {
	hname := must2(isc.HnameFromBytes([]byte(key[:4])))
	postfix := string(key[4:])
	return hname, postfix
}
