package main

import (
	"fmt"
	"log"
	"math/big"

	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/migrations/allmigrations"
	old_isc "github.com/nnikolash/wasp-types-exported/packages/isc"
	old_kv "github.com/nnikolash/wasp-types-exported/packages/kv"
	old_accounts "github.com/nnikolash/wasp-types-exported/packages/vm/core/accounts"
	old_allmigrations "github.com/nnikolash/wasp-types-exported/packages/vm/core/migrations/allmigrations"
	"github.com/samber/lo"
)

var (
	oldSchema old_isc.SchemaVersion = old_allmigrations.DefaultScheme.LatestSchemaVersion()
	newSchema isc.SchemaVersion     = allmigrations.SchemaVersionIotaRebased
)

type migratedAccount struct {
	OldAgentID old_isc.AgentID
	NewAgentID isc.AgentID
}

func migrateAccountsContract(oldChainState old_kv.KVStoreReader, newChainState state.StateDraft) {
	log.Print("Migrating accounts contract...\n")

	oldState := getContactStateReader(oldChainState, old_accounts.Contract.Hname())
	newState := getContactState(newChainState, accounts.Contract.Hname())

	oldChainID := old_isc.ChainID(GetAnchorOutput(oldChainState).AliasID)
	newChainID := isc.ChainID{} // TODO: Add as CLI argument

	migratedAccounts := map[old_kv.Key]migratedAccount{}

	migrateAccountsList(oldState, newState, oldChainID, newChainID, &migratedAccounts)
	migrateBaseTokenBalances(oldState, newState, oldChainID, newChainID, migratedAccounts)
	migrateNativeTokenBalances(oldState, newState, oldChainID, migratedAccounts)
	// migrateNativeTokenBalanceTotal(oldState, newState)
	// migrateFoundriesOutputs(oldState, newState)
	// migrateFoundriesPerAccount(oldState, newState, oldAgentIDToNewAgentID)
	// migrateNativeTokenOutputs(oldState, newState)
	// migrateAccountToNFT(oldState, newState, oldAgentIDToNewAgentID)
	// migrateNFTtoOwner(oldState, newState)
	// migrateNFTsByCollection(oldState, newState, oldAgentIDToNewAgentID)
	// prefixNewlyMintedNFTs ignored
	// migrateAllMintedNfts(oldState, newState)
	migrateNonce(oldState, newState, oldChainID, migratedAccounts)

	log.Print("Migrated accounts contract\n")
}

func migrateAccountsList(oldState old_kv.KVStoreReader, newState kv.KVStore, oldChID old_isc.ChainID, newChID isc.ChainID, migratedAccs *map[old_kv.Key]migratedAccount) {
	log.Printf("Migrating accounts list...\n")

	migrateAccountAndSaveNewAgentID := p(func(oldAccountKey old_kv.Key, v bool) (kv.Key, bool) {
		oldAgentID := lo.Must(old_accounts.AgentIDFromKey(oldAccountKey, oldChID))
		newAgentID := OldAgentIDtoNewAgentID(oldAgentID, newChID)

		(*migratedAccs)[oldAccountKey] = migratedAccount{
			OldAgentID: oldAgentID,
			NewAgentID: newAgentID,
		}

		return accounts.AccountKey(newAgentID, newChID), v
	})

	count := MigrateMapByName(
		oldState, newState,
		old_accounts.KeyAllAccounts, accounts.KeyAllAccounts,
		migrateAccountAndSaveNewAgentID,
	)

	log.Printf("Migrated list of %v accounts\n", count)
}

func migrateBaseTokenBalances(oldState old_kv.KVStoreReader, newState kv.KVStore, oldChainID old_isc.ChainID, newChainID isc.ChainID, migratedAccs map[old_kv.Key]migratedAccount) {
	log.Printf("Migrating base token balances...\n")

	for _, acc := range migratedAccs {
		oldBalance := old_accounts.GetBaseTokensBalanceFullDecimals(oldSchema, oldState, acc.OldAgentID, oldChainID)

		// NOTE: Next like affects also L2TotalsAccount, so its does not need to be migrated, only compared
		// TODO: What is the conversion rate here?
		w := accounts.NewStateWriter(newSchema, newState)
		w.CreditToAccountFullDecimals(acc.NewAgentID, oldBalance, newChainID)
	}

	log.Printf("Migrated %v base token balances\n", len(migratedAccs))
}

func migrateNativeTokenBalances(oldState old_kv.KVStoreReader, newState kv.KVStore, oldChainID old_isc.ChainID, migratedAccounts map[old_kv.Key]migratedAccount) {
	log.Printf("Migrating native token balances...\n")

	var count int

	for _, acc := range migratedAccounts {
		oldNativeTokes := old_accounts.GetNativeTokens(oldState, acc.OldAgentID, oldChainID)

		for _, oldNativeToken := range oldNativeTokes {
			if !oldNativeToken.Amount.IsUint64() {
				panic(fmt.Errorf("old native token amount cannot be represented as uint64: agentID = %v, token = %v, balance = %v",
					acc.OldAgentID, oldNativeToken.ID, oldNativeToken.Amount))
			}

			oldBalance := oldNativeToken.Amount.Uint64()

			newCoinType := OldNativeTokemIDtoNewCoinType(oldNativeToken.ID)
			newBalance := coin.Value(oldBalance) // TODO: What is the conversion rate?

			w := accounts.NewStateWriter(newSchema, newState)
			w.CreditToAccount(acc.NewAgentID, isc.CoinBalances{
				newCoinType: newBalance,
			}, isc.ChainID(oldChainID))
		}

		count += len(oldNativeTokes)
	}

	log.Printf("Migrated %v native token balances\n", count)
}

func migrateFoundriesOutputs(oldState old_kv.KVStoreReader, newState kv.KVStore) {
	panic("TODO: review")

	log.Printf("Migrating list of foundry outputs...\n")

	migrateEntry := func(oldKey old_kv.Key, oldVal old_accounts.FoundryOutputRec) (newKey kv.Key, newVal string) {
		return kv.Key(oldKey), "dummy new value"
	}

	count := MigrateMapByName(oldState, newState, old_accounts.KeyFoundryOutputRecords, "", p(migrateEntry))

	log.Printf("Migrated %v foundry outputs\n", count)
}

func migrateFoundriesPerAccount(oldState old_kv.KVStoreReader, newState kv.KVStore, oldAgentIDToNewAgentID map[old_isc.AgentID]isc.AgentID) {
	panic("TODO: review")

	log.Printf("Migrating foundries of accounts...\n")

	var count uint32

	migrateFoundriesOfAccount := p(func(oldKey old_kv.Key, oldVal bool) (newKey kv.Key, newVal bool) {
		return kv.Key(oldKey) + "dummy new key", oldVal
	})

	for oldAgentID, newAgentID := range oldAgentIDToNewAgentID {
		// mapMame := PrefixFoundries + string(agentID.Bytes())
		oldMapName := old_accounts.PrefixFoundries + string(oldAgentID.Bytes())
		_ = newAgentID
		newMapName := "" //accounts.PrefixFoundries + string(newAgentID.Bytes())

		count += MigrateMapByName(oldState, newState, oldMapName, newMapName, migrateFoundriesOfAccount)
	}

	log.Printf("Migrated %v foundries of accounts\n", count)
}

func migrateAccountToNFT(oldState old_kv.KVStoreReader, newState kv.KVStore, oldAgentIDToNewAgentID map[old_isc.AgentID]isc.AgentID) {
	panic("TODO: review")

	log.Printf("Migrating NFTs per account...\n")

	var count uint32
	migrateEntry := p(func(oldKey old_kv.Key, oldVal bool) (newKey kv.Key, newVal bool) {
		return kv.Key(oldKey) + "dummy new key", oldVal
	})

	for oldAgentID, newAgentID := range oldAgentIDToNewAgentID {
		// mapName := PrefixNFTs + string(agentID.Bytes())
		oldMapName := old_accounts.PrefixNFTs + string(oldAgentID.Bytes())
		_ = newAgentID
		newMapName := "" // accounts.PrefixNFTs + string(newAgentID.Bytes())

		count += MigrateMapByName(oldState, newState, oldMapName, newMapName, migrateEntry)
	}

	log.Printf("Migrated %v NFTs per account\n", count)
}

func migrateNFTtoOwner(oldState old_kv.KVStoreReader, newState kv.KVStore) {
	panic("TODO: review")

	log.Printf("Migrating NFT owners...\n")

	migrateEntry := func(oldKey old_kv.Key, oldVal []byte) (newKey kv.Key, newVal []byte) {
		return kv.Key(oldKey) + "dummy new key", append(oldVal, []byte("dummy new value")...)
	}

	count := MigrateMapByName(oldState, newState, old_accounts.KeyNFTOwner, "", p(migrateEntry))
	log.Printf("Migrated %v NFT owners\n", count)
}

func migrateNFTsByCollection(oldState old_kv.KVStoreReader, newState kv.KVStore, oldAgentIDToNewAgentID map[old_isc.AgentID]isc.AgentID) {
	panic("TODO: review")

	log.Printf("Migrating NFTs by collection...\n")

	var count uint32

	for oldAgentID, newAgentID := range oldAgentIDToNewAgentID {
		// mapName := PrefixNFTsByCollection + string(agentID.Bytes()) + string(collectionID.Bytes())
		// NOTE: There is no easy way to retrieve list of referenced collections
		oldPrefix := old_accounts.PrefixNFTsByCollection + string(oldAgentID.Bytes())
		newPrefix := "" // accounts.PrefixNFTsByCollection + string(newAgentID.Bytes())

		count += MigrateByPrefix(oldState, newState, oldPrefix, newPrefix, func(oldKey old_kv.Key, oldVal bool) (newKey kv.Key, newVal bool) {
			return migrateNFTsByCollectionEntry(oldKey, oldVal, oldAgentID, newAgentID)
		})
	}

	log.Printf("Migrated %v NFTs by collection\n", count)
}

func migrateNFTsByCollectionEntry(oldKey old_kv.Key, oldVal bool, oldAgentID old_isc.AgentID, newAgentID isc.AgentID) (newKey kv.Key, newVal bool) {
	panic("TODO: review")

	oldMapName, oldMapElemKey := SplitMapKey(oldKey)

	oldPrefix := old_accounts.PrefixNFTsByCollection + string(oldAgentID.Bytes())
	collectionIDBytes := oldMapName[len(oldPrefix):]
	_ = collectionIDBytes

	var newMapName kv.Key = "" // accounts.PrefixNFTsByCollection + string(newAgentID.Bytes()) + string(collectionIDBytes)

	newKey = newMapName

	if oldMapElemKey != "" {
		// If this record is map element - we form map element key.
		nftID := oldMapElemKey
		// TODO: migrate NFT ID
		newKey += "." + kv.Key(nftID)
	}

	return kv.Key(newKey), oldVal
}

func migrateNativeTokenOutputs(oldState old_kv.KVStoreReader, newState kv.KVStore) {
	panic("TODO: review")

	log.Printf("Migrating native token outputs...\n")

	migrateEntry := func(oldKey old_kv.Key, oldVal old_accounts.NativeTokenOutputRec) (newKey kv.Key, newVal old_accounts.NativeTokenOutputRec) {
		return kv.Key(oldKey), oldVal
	}

	count := MigrateMapByName(oldState, newState, old_accounts.KeyNativeTokenOutputMap, "", p(migrateEntry))

	log.Printf("Migrated %v native token outputs\n", count)
}

func migrateNativeTokenBalanceTotal(oldState old_kv.KVStoreReader, newState kv.KVStore) {
	panic("TODO: review")

	log.Printf("Migrating native token total balance...\n")

	migrateEntry := func(oldKey old_kv.Key, oldVal *big.Int) (newKey kv.Key, newVal []byte) {
		// TODO: new amount format (if not big.Int)
		return kv.Key(oldKey), []byte{0}
	}

	count := MigrateMapByName(oldState, newState, old_accounts.PrefixNativeTokens+accounts.L2TotalsAccount, "", p(migrateEntry))

	log.Printf("Migrated %v native token total balance\n", count)
}

func migrateAllMintedNfts(oldState old_kv.KVStoreReader, newState kv.KVStore) {
	panic("TODO: review")

	// prefixMintIDMap stores a map of <internal NFTID> => <NFTID>
	log.Printf("Migrating All minted NFTs...\n")

	migrateEntry := func(oldKey old_kv.Key, oldVal []byte) (newKey kv.Key, newVal []byte) {
		return kv.Key(oldKey), []byte{0}
	}

	count := MigrateMapByName(oldState, newState, old_accounts.PrefixMintIDMap, "", p(migrateEntry))

	log.Printf("Migrated %v All minted NFTs\n", count)
}

func migrateNonce(oldState old_kv.KVStoreReader, newState kv.KVStore, oldChainID old_isc.ChainID, migratedAccounts map[old_kv.Key]migratedAccount) {
	log.Printf("Migrating nonce...\n")

	for _, acc := range migratedAccounts {
		nonce := old_accounts.AccountNonce(oldState, acc.OldAgentID, oldChainID)
		setAccountNonce(newState, acc.NewAgentID, isc.ChainID(oldChainID), nonce)
	}

	log.Printf("Migrated %v nonce\n", len(migratedAccounts))
}
