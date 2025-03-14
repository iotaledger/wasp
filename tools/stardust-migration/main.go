// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"log"
	"math/big"
	"os"
	"runtime"

	hivedb "github.com/iotaledger/hive.go/db"
	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/hive.go/kvstore/rocksdb"
	"github.com/iotaledger/wasp/packages/database"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/subrealm"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/state/indexedstore"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
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
	migrateOtherContractStates(srcState, destStateDraft)

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

func getContactStateReader(chainState kv.KVStoreReader, contractHname isc.Hname) kv.KVStoreReader {
	return subrealm.NewReadOnly(chainState, kv.Key(contractHname.Bytes()))
}

func getContactState(chainState kv.KVStore, contractHname isc.Hname) kv.KVStore {
	return subrealm.New(chainState, kv.Key(contractHname.Bytes()))
}

func migrateAccountsContractState2(srcChainState kv.KVStoreReader, destChainState state.StateDraft) {
	srcContractState := getContactStateReader(srcChainState, coreutil.CoreHname(accounts.Contract.Name))
	destContractState := getContactState(destChainState, coreutil.CoreHname(accounts.Contract.Name))

	chainID := isc.ChainID(GetAnchorOutput(srcChainState).AliasID)

	log.Print("Migrating accounts contract state...\n")

	// Accounts
	log.Printf("Migrating accounts...\n")
	oldAgentIDToNewAgentID := map[isc.AgentID]isc.AgentID{}

	count := migrateEntitiesMapByName(srcContractState, destContractState, keyAllAccounts, "", p(func(oldAccountKey kv.Key, srcVal bool) (kv.Key, bool) {
		oldAgentID := must2(accounts.AgentIDFromKey(oldAccountKey, chainID))
		newAgentID, newV := migrateAccount(oldAgentID, srcVal)
		oldAgentIDToNewAgentID[oldAgentID] = newAgentID

		return accountKey(newAgentID, chainID), newV
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
		oldMapName := PrefixFoundries + string(oldAgentID.Bytes())
		newMapName := PrefixFoundries + string(newAgentID.Bytes())

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
		oldMapName := PrefixNativeTokens + string(accountKey(oldAgentID, chainID))
		newMapName := PrefixNativeTokens + string(accountKey(newAgentID, chainID))
		count += migrateEntitiesMapByName(srcContractState, destContractState, oldMapName, newMapName, migrateNativeTokenBalance)
	}

	log.Printf("Migrated %v native token balances\n", count)

	// Account to NFT
	// mapName := PrefixNFTs + string(agentID.Bytes())
	log.Printf("Migrating NFTs per account...\n")

	count = 0
	migrateNFT := p(migrateNFT)
	for oldAgentID, newAgentID := range oldAgentIDToNewAgentID {
		oldMapName := PrefixNFTs + string(oldAgentID.Bytes())
		newMapName := PrefixNFTs + string(newAgentID.Bytes())
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
		oldPrefix := PrefixNFTsByCollection + string(oldAgentID.Bytes())
		count += migrateEntitiesByPrefix(srcContractState, destContractState, oldPrefix, func(oldKey kv.Key, srcVal bool) (destKey kv.Key, destVal bool) {
			return migrateNFTByCollection(oldKey, srcVal, oldAgentID, newAgentID)
		})
	}

	log.Printf("Migrated %v NFTs by collection\n", count)

	// Native token outputs
	log.Printf("Migrating native token outputs...\n")

	count = migrateEntitiesMapByName(srcContractState, destContractState, keyNativeTokenOutputMap, "", p(migrateNativeTokenOutput))

	log.Printf("Migrated %v native token outputs\n", count)

	// Native token total balance
	log.Printf("Migrating native token total balance...\n")

	count = migrateEntitiesMapByName(srcContractState, destContractState, PrefixNativeTokens+L2TotalsAccount, "", p(migrateNativeTokenBalanceTotal))

	log.Printf("Migrated %v native token total balance\n", count)

	// All minted NFTs
	// prefixMintIDMap stores a map of <internal NFTID> => <NFTID>
	log.Printf("Migrating All minted NFTs...\n")

	count = migrateEntitiesMapByName(srcContractState, destContractState, prefixMintIDMap, "", p(migrateAllMintedNfts))

	log.Printf("Migrated %v All minted NFTs\n", count)

	log.Print("Migrated accounts contract state\n")
}

func migrateOtherContractStates(srcChainState kv.KVStoreReader, destChainState state.StateDraft) {

	//srcContractState := getContactStateReader(srcChainState, coreutil.CoreHname(blocklog.Contract.Name))
	// destContractState := getContactState(destChainState, coreutil.CoreHname(accounts.Contract.Name))

	governanceContractStateSrc := getContactStateReader(srcChainState, coreutil.CoreHname(governance.Contract.Name))
	governanceContractStateDest := getContactState(destChainState, coreutil.CoreHname(governance.Contract.Name))

	log.Print("Migrating other contracts states\n")

	// Unprocessable Requests (blocklog contract)
	// No need to migrate. Just print a warning if there are any
	log.Printf("Listing Unprocessable Requests...\n")

	count := 0
	collections.NewMapReadOnly(getContactStateReader(srcChainState, coreutil.CoreHname(blocklog.Contract.Name)), prefixUnprocessableRequests).Iterate(func(srcKey, srcBytes []byte) bool {
		reqID := must2(DeserializeEntity[isc.RequestID](srcKey))
		log.Printf("Warning: unprocessable request found %v", reqID.String())
		count++
		return true
	})
	log.Printf("Listing Unprocessable Requests completed (found %v entities)\n", count)

	// Chain Owner
	log.Printf("Migrating chain owner...\n")
	migrateEntityState(governanceContractStateSrc, governanceContractStateDest, governance.VarChainOwnerID, migrateAsIs)
	log.Printf("Migrated chain owner\n")

	// Chain Owner delegated
	log.Printf("Migrating chain owner delegated...\n")
	migrateEntityState(governanceContractStateSrc, governanceContractStateDest, governance.VarChainOwnerIDDelegated, migrateAsIs)
	log.Printf("Migrated chain owner delegated\n")

	// Payout agent
	log.Printf("Migrating Payout agent...\n")
	migrateEntityState(governanceContractStateSrc, governanceContractStateDest, governance.VarPayoutAgentID, migrateAsIs)
	log.Printf("Migrated Payout agent\n")

	// Min Base Tokens On Common Account
	log.Printf("Migrating Min Base Tokens On Common Account...\n")
	migrateEntityState(governanceContractStateSrc, governanceContractStateDest, governance.VarMinBaseTokensOnCommonAccount, migrateAsIs)
	log.Printf("Migrated Min Base Tokens On Common Account\n")

	log.Print("Migrated other contracts states\n")
}

func migrateAsIs(srcKey kv.Key, srcVal []byte) (destKey kv.Key, destVal []byte) {
	return srcKey, srcVal
}

func migrateAccount(oldAgentID isc.AgentID, srcVal bool) (newAgentID isc.AgentID, destVal bool) {
	return oldAgentID, srcVal
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

func migrateNFTByCollection(oldKey kv.Key, srcVal bool, oldAgentID, newAgentID isc.AgentID) (destKey kv.Key, destVal bool) {
	oldMapName, oldMapElemKey := SplitMapKey(oldKey)

	oldPrefix := PrefixNFTsByCollection + string(oldAgentID.Bytes())
	collectionIDBytes := oldMapName[len(oldPrefix):]

	newMapName := PrefixNFTsByCollection + string(newAgentID.Bytes()) + string(collectionIDBytes)

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

func migrateNativeTokenBalanceTotal(srcKey kv.Key, srcVal *big.Int) (destKey kv.Key, destVal []byte) {
	// TODO: new amount format (if not big.Int)
	return srcKey, []byte{0}
}

func migrateAllMintedNfts(srcKey kv.Key, srcVal []byte) (destKey kv.Key, destVal []byte) {
	return srcKey, []byte{0}
}
