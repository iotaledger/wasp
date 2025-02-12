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
	"github.com/iotaledger/wasp/packages/coin"
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

func MigrateAccountsContract(
	v old_isc.SchemaVersion,
	oldChainState old_kv.KVStoreReader,
	newChainState state.StateDraft,
	oldChainID old_isc.ChainID,
	newChainID isc.ChainID,
) {
	log.Print("Migrating accounts contract...\n")
	oldState := oldstate.GetContactStateReader(oldChainState, old_accounts.Contract.Hname())
	newState := newstate.GetContactState(newChainState, accounts.Contract.Hname())

	migratedAccounts := map[old_kv.Key]migratedAccount{}

	migrateAccountsList(oldState, newState, oldChainID, newChainID, &migratedAccounts)
	migrateBaseTokenBalances(v, oldState, newState, oldChainID, newChainID, migratedAccounts)
	migrateNativeTokenBalances(oldState, newState, oldChainID, newChainID, migratedAccounts)
	// NOTE: L2TotalsAccount is migrated implicitly inside of migrateBaseTokenBalances and migrateNativeTokenBalances
	migrateFoundriesOutputs(oldState, newState)
	migrateFoundriesPerAccount(oldState, newState, oldChainID, newChainID)
	migrateNativeTokenOutputs(oldState, newState)
	migrateNFTs(oldState, newState, oldChainID, newChainID)
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

func migrateBaseTokenBalances(
	v old_isc.SchemaVersion,
	oldState old_kv.KVStoreReader,
	newState kv.KVStore,
	oldChainID old_isc.ChainID,
	newChainID isc.ChainID,
	migratedAccs map[old_kv.Key]migratedAccount,
) {
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
	// new: foundries not supported, just backup the map

	count := BackupMapByName(
		oldState,
		newState,
		old_accounts.KeyFoundryOutputRecords,
	)

	log.Printf("Migrated %v foundry outputs\n", count)
}

func migrateFoundriesPerAccount(
	oldState old_kv.KVStoreReader,
	newState kv.KVStore,
	oldChainID old_isc.ChainID,
	newChainID isc.ChainID,
) {
	log.Printf("Migrating foundries of accounts...\n")

	// old: PrefixFoundries + <agentID> stores a map of <foundrySN> (uint32) => true
	// new: foundries not supported, just backup the maps

	const sizeofFoundrySN = 4
	count := BackupAccountMaps(
		oldState,
		newState,
		old_accounts.PrefixFoundries,
		sizeofFoundrySN,
		oldChainID,
		newChainID,
	)

	log.Printf("Migrated %v foundries of accounts\n", count)
}

func migrateNativeTokenOutputs(oldState old_kv.KVStoreReader, newState kv.KVStore) {
	log.Printf("Migrating native token outputs...\n")

	migrateEntry := func(ntID old_iotago.NativeTokenID, rec old_accounts.NativeTokenOutputRec) (coin.Type, isc.IotaCoinInfo) {
		coinType := OldNativeTokenIDtoNewCoinType(ntID)
		coinInfo := OldNativeTokenIDtoNewCoinInfo(ntID)
		return coinType, coinInfo
	}

	// old: KeyNativeTokenOutputMap stores a map of <nativeTokenID> => nativeTokenOutputRec
	// new: keyCoinInfo ("RC") stores a map of <CoinType> => isc.IotaCoinInfo
	count := MigrateMapByName(oldState, newState, old_accounts.KeyNativeTokenOutputMap, "RC", p(migrateEntry))

	log.Printf("Migrated %v native token outputs\n", count)
}

func migrateNFTs(
	oldState old_kv.KVStoreReader,
	newState kv.KVStore,
	oldChainID old_isc.ChainID,
	newChainID isc.ChainID,
) {
	log.Printf("Migrating NFTs...\n")

	oldNFTOutputs := old_accounts.NftOutputMapR(oldState)
	w := accounts.NewStateWriter(newSchema, newState)

	var count uint32
	oldNFTOutputs.IterateKeys(func(k []byte) bool {
		nftID := old_codec.MustDecodeNFTID([]byte(k))
		oldNFT := old_accounts.GetNFTData(oldState, nftID)
		owner := OldAgentIDtoNewAgentID(oldNFT.Owner, oldChainID, newChainID)
		newObjectRecord := OldNFTIDtoNewObjectRecord(nftID)
		w.CreditObjectToAccount(owner, newObjectRecord, newChainID)
		count++
		return true
	})

	log.Printf("Migrated %v NFTs\n", count)
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
