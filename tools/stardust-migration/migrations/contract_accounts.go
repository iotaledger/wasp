package migrations

import (
	"log"
	"math/big"

	old_isc "github.com/nnikolash/wasp-types-exported/packages/isc"
	old_kv "github.com/nnikolash/wasp-types-exported/packages/kv"
	old_codec "github.com/nnikolash/wasp-types-exported/packages/kv/codec"
	old_accounts "github.com/nnikolash/wasp-types-exported/packages/vm/core/accounts"
	"github.com/samber/lo"

	old_iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/stardust-migration/stateaccess/newstate"
	"github.com/iotaledger/stardust-migration/stateaccess/oldstate"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
)

type migratedAccount struct {
	OldAgentID old_isc.AgentID
	NewAgentID isc.AgentID
}

func MigrateAccountsContract(v old_isc.SchemaVersion, oldChainState old_kv.KVStoreReader, newChainState state.StateDraft, oldChainID old_isc.ChainID, newChainID isc.ChainID) {
	log.Print("Migrating accounts contract...\n")
	oldState := oldstate.GetContactStateReader(oldChainState, old_accounts.Contract.Hname())
	newState := newstate.GetContactState(newChainState, accounts.Contract.Hname())

	migratedAccounts := map[old_kv.Key]migratedAccount{}

	migrateAccountsList(oldState, newState, oldChainID, newChainID, &migratedAccounts)
	migrateBaseTokenBalances(v, oldState, newState, oldChainID, newChainID, migratedAccounts)
	migrateNativeTokenBalances(oldState, newState, oldChainID, newChainID, migratedAccounts)
	// NOTE: L2TotalsAccount is migrated implicitly inside of migrateBaseTokenBalances and migrateNativeTokenBalances
	migrateFoundriesOutputs(oldState, newState)
	migrateFoundriesPerAccount(oldState, newState)
	migrateNativeTokenOutputs(oldState, newState)
	migrateAccountToNFT(oldState, newState)
	migrateNFTtoOwner(oldState, newState)
	// migrateNFTsByCollection(oldState, newState)
	// prefixNewlyMintedNFTs ignored
	// migrateAllMintedNfts(oldState, newState)
	migrateNonce(oldState, newState, oldChainID, newChainID, migratedAccounts)

	log.Print("Migrated accounts contract\n")
}

func migrateAccountsList(oldState old_kv.KVStoreReader, newState kv.KVStore, oldChID old_isc.ChainID, newChID isc.ChainID, migratedAccs *map[old_kv.Key]migratedAccount) {
	log.Printf("Migrating accounts list...\n")

	migrateAccountAndSaveNewAgentID := p(func(oldAccountKey old_kv.Key, v bool) (kv.Key, bool) {
		oldAgentID := lo.Must(old_accounts.AgentIDFromKey(oldAccountKey, oldChID))
		newAgentID := OldAgentIDtoNewAgentID(oldAgentID, oldChID, newChID)

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

func convertBaseTokens(oldBalanceFullDecimals *big.Int) *big.Int {
	panic("TODO: do we need to apply a conversion rate because of iota's 6 to 9 decimals change?")
	return oldBalanceFullDecimals
}

func migrateBaseTokenBalances(v old_isc.SchemaVersion, oldState old_kv.KVStoreReader, newState kv.KVStore, oldChainID old_isc.ChainID, newChainID isc.ChainID, migratedAccs map[old_kv.Key]migratedAccount) {
	log.Printf("Migrating base token balances...\n")

	w := accounts.NewStateWriter(newSchema, newState)
	for _, acc := range migratedAccs {
		oldBalance := old_accounts.GetBaseTokensBalanceFullDecimals(v, oldState, acc.OldAgentID, oldChainID)

		// NOTE: L2TotalsAccount is also credited here, so it does not need to be migrated, only compared.
		w.CreditToAccountFullDecimals(acc.NewAgentID, convertBaseTokens(oldBalance), newChainID)
	}

	log.Printf("Migrated %v base token balances\n", len(migratedAccs))
}

func migrateNativeTokenBalances(oldState old_kv.KVStoreReader, newState kv.KVStore, oldChainID old_isc.ChainID, newChainID isc.ChainID, migratedAccounts map[old_kv.Key]migratedAccount) {
	log.Printf("Migrating native token balances...\n")

	var count int

	w := accounts.NewStateWriter(newSchema, newState)
	for _, acc := range migratedAccounts {
		oldNativeTokes := old_accounts.GetNativeTokens(oldState, acc.OldAgentID, oldChainID)

		for _, oldNativeToken := range oldNativeTokes {
			newCoinType := OldNativeTokenIDtoNewCoinType(oldNativeToken.ID)
			newBalance := OldNativeTokenBalanceToNewCoinValue(oldNativeToken.Amount)

			// NOTE: L2TotalsAccount is also credited here, so it does not need to be migrated, only compared.
			w.CreditToAccount(acc.NewAgentID, isc.CoinBalances{newCoinType: newBalance}, newChainID)
		}

		count += len(oldNativeTokes)
	}

	log.Printf("Migrated %v native token balances\n", count)
}

func migrateFoundriesOutputs(oldState old_kv.KVStoreReader, newState kv.KVStore) {
	log.Printf("Migrating list of foundry outputs...\n")

	// old: KeyFoundryOutputRecords stores a map of <foundrySN> => foundryOutputRec
	// new: ??
	migrateEntry := func(oldKey old_kv.Key, oldVal old_accounts.FoundryOutputRec) (newKey kv.Key, newVal string) {
		panic("TODO: what should we do with foundries?")
	}

	count := MigrateMapByName(oldState, newState, old_accounts.KeyFoundryOutputRecords, "", p(migrateEntry))

	log.Printf("Migrated %v foundry outputs\n", count)
}

func migrateFoundriesPerAccount(oldState old_kv.KVStoreReader, newState kv.KVStore) {
	log.Printf("Migrating foundries of accounts...\n")

	// old: PrefixFoundries + <agentID> stores a map of <foundrySN> (uint32) => true
	// new: ??

	const sizeofFoundrySN = 4
	count := oldstate.IterateAccountMaps(oldState, old_accounts.PrefixFoundries, sizeofFoundrySN, func(oldAgentID *old_isc.AgentID, mapKey old_kv.Key, _ []byte) {
		foundrySN := old_codec.MustDecodeUint32([]byte(mapKey))
		_ = foundrySN
		panic("TODO: what should we do with foundries?")
	})

	log.Printf("Migrated %v foundries of accounts\n", count)
}

func migrateAccountToNFT(oldState old_kv.KVStoreReader, newState kv.KVStore) {
	log.Printf("Migrating NFTs per account...\n")

	// old: PrefixNFTs | <agentID> stores a map of <NFTID> => true
	// new: prefixObjects "o" | <agentID> stores a map of <ObjectID> => true
	count := oldstate.IterateAccountMaps(oldState, old_accounts.PrefixNFTs, old_iotago.NFTIDLength, func(oldAgentID *old_isc.AgentID, mapKey old_kv.Key, _ []byte) {
		nftID := old_codec.MustDecodeNFTID([]byte(mapKey))
		_ = nftID
		panic("TODO: what should we do with NFTs?")
	})

	log.Printf("Migrated %v NFTs per account\n", count)
}

func migrateNFTtoOwner(oldState old_kv.KVStoreReader, newState kv.KVStore) {
	log.Printf("Migrating NFT owners...\n")

	// old: KeyNFTOwner stores a map of <NFTID> => isc.AgentID
	// new: keyObjectOwner "W" stores a map of <ObjectID> => isc.AgentID
	migrateEntry := func(oldKey old_kv.Key, oldVal []byte) (newKey kv.Key, newVal []byte) {
		panic("TODO: what should we do with NFTs?")
	}

	count := MigrateMapByName(oldState, newState, old_accounts.KeyNFTOwner, "", p(migrateEntry))
	log.Printf("Migrated %v NFT owners\n", count)
}

func migrateNFTsByCollection(oldState old_kv.KVStoreReader, newState kv.KVStore) {
	panic("TODO: implement (using existing business logic)")

	log.Printf("Migrating NFTs by collection...\n")

	var count uint32

	// for oldAgentID, newAgentID := range oldAgentIDToNewAgentID {
	// 	// mapName := PrefixNFTsByCollection + string(agentID.Bytes()) + string(collectionID.Bytes())
	// 	// NOTE: There is no easy way to retrieve list of referenced collections
	// 	oldPrefix := old_accounts.PrefixNFTsByCollection + string(oldAgentID.Bytes())
	// 	newPrefix := "" // accounts.PrefixNFTsByCollection + string(newAgentID.Bytes())
	//
	// 	count += MigrateByPrefix(oldState, newState, oldPrefix, newPrefix, func(oldKey old_kv.Key, oldVal bool) (newKey kv.Key, newVal bool) {
	// 		return migrateNFTsByCollectionEntry(oldKey, oldVal, oldAgentID, newAgentID)
	// 	})
	// }
	//
	log.Printf("Migrated %v NFTs by collection\n", count)
}

func migrateNFTsByCollectionEntry(oldKey old_kv.Key, oldVal bool, oldAgentID old_isc.AgentID, newAgentID isc.AgentID) (newKey kv.Key, newVal bool) {
	panic("TODO: implement (using existing business logic)")

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
	log.Printf("Migrating native token outputs...\n")

	migrateEntry := func(oldKey old_kv.Key, oldVal old_accounts.NativeTokenOutputRec) (newKey kv.Key, newVal isc.IotaCoinInfo) {
		panic("TODO: how to migrate native tokens => coins")
	}

	// old: KeyNativeTokenOutputMap stores a map of <nativeTokenID> => nativeTokenOutputRec
	// new: keyCoinInfo ("RC") stores a map of <CoinType> => isc.IotaCoinInfo
	count := MigrateMapByName(oldState, newState, old_accounts.KeyNativeTokenOutputMap, "RC", p(migrateEntry))

	log.Printf("Migrated %v native token outputs\n", count)
}

func migrateAllMintedNfts(oldState old_kv.KVStoreReader, newState kv.KVStore) {
	panic("TODO: implement (using existing business logic)")

	// prefixMintIDMap stores a map of <internal NFTID> => <NFTID>
	log.Printf("Migrating All minted NFTs...\n")

	migrateEntry := func(oldKey old_kv.Key, oldVal []byte) (newKey kv.Key, newVal []byte) {
		return kv.Key(oldKey), []byte{0}
	}

	count := MigrateMapByName(oldState, newState, old_accounts.PrefixMintIDMap, "", p(migrateEntry))

	log.Printf("Migrated %v All minted NFTs\n", count)
}

func migrateNonce(oldState old_kv.KVStoreReader, newState kv.KVStore, oldChainID old_isc.ChainID, newChainID isc.ChainID, migratedAccounts map[old_kv.Key]migratedAccount) {
	log.Printf("Migrating nonce...\n")

	for _, acc := range migratedAccounts {
		if acc.NewAgentID.Kind() == isc.AgentIDKindEthereumAddress {
			// don't update EVM nonces
			return
		}

		nonce := old_accounts.AccountNonce(oldState, acc.OldAgentID, oldChainID)
		newState.Set(accounts.NonceKey(acc.NewAgentID, newChainID), codec.Encode(nonce))
	}

	log.Printf("Migrated %v nonce\n", len(migratedAccounts))
}
