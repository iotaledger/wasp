package migrations

import (
	"math/big"

	"github.com/samber/lo"

	old_iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/parameters"

	old_isc "github.com/nnikolash/wasp-types-exported/packages/isc"
	old_kv "github.com/nnikolash/wasp-types-exported/packages/kv"
	old_codec "github.com/nnikolash/wasp-types-exported/packages/kv/codec"
	old_parameters "github.com/nnikolash/wasp-types-exported/packages/parameters"
	old_util "github.com/nnikolash/wasp-types-exported/packages/util"
	old_accounts "github.com/nnikolash/wasp-types-exported/packages/vm/core/accounts"

	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"

	"github.com/iotaledger/wasp/tools/stardust-migration/stateaccess/newstate"
	"github.com/iotaledger/wasp/tools/stardust-migration/stateaccess/oldstate"
	"github.com/iotaledger/wasp/tools/stardust-migration/utils"
	"github.com/iotaledger/wasp/tools/stardust-migration/utils/cli"
)

type migratedAccount struct {
	OldAgentID old_isc.AgentID
	NewAgentID isc.AgentID
}

func MigrateAccountsContract(
	v old_isc.SchemaVersion,
	prevChainState old_kv.KVStoreReader,
	oldChainState old_kv.KVStoreReader,
	newChainState kv.KVStore,
	oldChainID old_isc.ChainID,
) {
	cli.DebugLog("Migrating accounts contract (muts)...\n")
	prevOldState := oldstate.GetContactStateReader(prevChainState, old_accounts.Contract.Hname())
	oldState := oldstate.GetContactStateReader(oldChainState, old_accounts.Contract.Hname())
	newState := newstate.GetContactState(newChainState, accounts.Contract.Hname())

	migrateAccountsList(oldState, newState, oldChainID)
	migrateBaseTokenBalances(v, oldState, newState, oldChainID)
	migrateNativeTokenBalances(oldState, newState, oldChainID)
	migrateFoundriesOutputs(oldState, newState)
	migrateFoundriesPerAccount(oldState, newState)
	migrateNativeTokenOutputs(oldState, newState)
	migrateNFTs(prevOldState, oldState, newState, oldChainID)
	// prefixNewlyMintedNFTs ignored
	// PrefixMintIDMap ignored
	migrateNonce(oldState, newState, oldChainID)

	cli.DebugLog("Migrated accounts contract (muts)\n")
}

func migrateAccountsList(oldState old_kv.KVStoreReader, newState kv.KVStore, oldChID old_isc.ChainID) {
	cli.DebugLogf("Migrating accounts list...\n")

	count := MigrateMapByName(
		oldState, newState,
		old_accounts.KeyAllAccounts, accounts.KeyAllAccounts,
		func(oldAccountKey old_kv.Key, v *bool) (kv.Key, *bool) {
			oldAgentID := lo.Must(old_accounts.AgentIDFromKey(oldAccountKey, oldChID))
			newAgentID := OldAgentIDtoNewAgentID(oldAgentID, oldChID)
			return accounts.AccountKey(newAgentID), v
		},
	)

	cli.DebugLogf("Migrated list of %v accounts\n", count)
}

func migrateBaseTokenBalances(
	v old_isc.SchemaVersion,
	oldState old_kv.KVStoreReader,
	newState kv.KVStore,
	oldChainID old_isc.ChainID,
) {
	cli.DebugLogf("Migrating base token balances...\n")

	w := accounts.NewStateWriter(newSchema, newState)
	count := 0

	oldState.Iterate(old_accounts.PrefixBaseTokens, func(k old_kv.Key, oldBalanceBytes []byte) bool {
		count++
		oldAccKey := utils.MustRemovePrefix(k, old_accounts.PrefixBaseTokens)

		var newAccKey kv.Key
		if oldAccKey == old_accounts.L2TotalsAccount {
			newAccKey = accounts.L2TotalsAccount
		} else {
			oldAgentID := lo.Must(old_accounts.AgentIDFromKey(old_kv.Key(oldAccKey), oldChainID))
			newAgentID := OldAgentIDtoNewAgentID(oldAgentID, oldChainID)
			newAccKey = accounts.AccountKey(newAgentID)
		}

		var balance *big.Int
		if oldBalanceBytes != nil {
			switch v {
			case 0:
				amount := old_codec.MustDecodeUint64(oldBalanceBytes, 0)
				balance = old_util.BaseTokensDecimalsToEthereumDecimals(amount, old_parameters.L1().BaseToken.Decimals)
			default:
				balance = old_codec.MustDecodeBigIntAbs(oldBalanceBytes, big.NewInt(0))
			}

			// NOTE: We do NOT need to apply conversion here - full decimal value stays same,
			// because number of digits has changes for internal representation, but not for ethereum.
		}

		w.UnsafeSetBaseTokensFullDecimals(newAccKey, balance)

		return true
	})

	cli.DebugLogf("Migrated %v base token balances\n", count)
}

func migrateNativeTokenBalances(oldState old_kv.KVStoreReader, newState kv.KVStore, oldChainID old_isc.ChainID) {
	cli.DebugLogf("Migrating native token balances...\n")

	var count int

	w := accounts.NewStateWriter(newSchema, newState)
	oldState.Iterate(old_accounts.PrefixNativeTokens, func(k old_kv.Key, v []byte) bool {
		count++
		oldAccKey, oldNtIDBytes := utils.MustSplitMapKey(k, -old_iotago.FoundryIDLength-1, old_accounts.PrefixNativeTokens)
		if oldNtIDBytes == "" {
			// not a map entry
			return true
		}

		var newAccKey kv.Key
		if oldAccKey == old_accounts.L2TotalsAccount {
			newAccKey = accounts.L2TotalsAccount
		} else {
			oldAgentID := lo.Must(old_accounts.AgentIDFromKey(old_kv.Key(oldAccKey), oldChainID))
			newAgentID := OldAgentIDtoNewAgentID(oldAgentID, oldChainID)
			newAccKey = accounts.AccountKey(newAgentID)
		}

		oldNtID := old_isc.MustNativeTokenIDFromBytes([]byte(oldNtIDBytes))
		newCoinType := OldNativeTokenIDtoNewCoinType(oldNtID)

		var newBalance coin.Value
		if v != nil {
			oldBalance := old_codec.MustDecodeBigIntAbs(v, new(big.Int))
			newBalance = OldNativeTokenBalanceToNewCoinValue(oldBalance)
		}

		w.UnsafeSetCoinBalance(newAccKey, newCoinType, newBalance)

		return true
	})

	cli.DebugLogf("Migrated %v native token balances\n", count)
}

func migrateFoundriesOutputs(oldState old_kv.KVStoreReader, newState kv.KVStore) {
	cli.DebugLogf("Migrating list of foundry outputs...\n")

	// old: KeyFoundryOutputRecords stores a map of <foundrySN> => foundryOutputRec
	// new: foundries not supported, just backup the map

	count := BackupMapByName(
		oldState,
		newState,
		old_accounts.KeyFoundryOutputRecords,
	)

	cli.DebugLogf("Migrated %v foundry outputs\n", count)
}

func migrateFoundriesPerAccount(
	oldState old_kv.KVStoreReader,
	newState kv.KVStore,
) {
	cli.DebugLogf("Migrating foundries of accounts...\n")

	// old: PrefixFoundries + <agentID> stores a map of <foundrySN> (uint32) => true
	// new: foundries not supported, just backup the maps

	count := BackupByPrefix(oldState, newState, old_accounts.PrefixFoundries)
	cli.DebugLogf("Migrated %v foundries of accounts\n", count)
}

func migrateNativeTokenOutputs(oldState old_kv.KVStoreReader, newState kv.KVStore) {
	cli.DebugLogf("Migrating native token outputs...\n")

	migrateEntry := func(ntID old_iotago.NativeTokenID, rec *old_accounts.NativeTokenOutputRec) (coin.Type, *parameters.IotaCoinInfo) {
		coinType := OldNativeTokenIDtoNewCoinType(ntID)
		if rec == nil {
			return coinType, nil
		}

		coinInfo := OldNativeTokenIDtoNewCoinInfo(ntID)
		return coinType, &coinInfo
	}

	// old: KeyNativeTokenOutputMap stores a map of <nativeTokenID> => nativeTokenOutputRec
	// new: keyCoinInfo ("RC") stores a map of <CoinType> => isc.IotaCoinInfo
	count := MigrateMapByName(oldState, newState, old_accounts.KeyNativeTokenOutputMap, "RC", p(migrateEntry))

	cli.DebugLogf("Migrated %v native token outputs\n", count)
}

func migrateNFTs(
	prevOldState old_kv.KVStoreReader,
	oldState old_kv.KVStoreReader,
	newState kv.KVStore,
	oldChainID old_isc.ChainID,
) {
	cli.DebugLogf("Migrating NFTs...\n")

	prevOldNFTToOwner := old_accounts.NftToOwnerMapR(prevOldState)
	oldNFTToOwner := old_accounts.NftToOwnerMapR(oldState)
	w := accounts.NewStateWriter(newSchema, newState)

	var count uint32
	oldNFTToOwner.Iterate(func(k, v []byte) bool {
		nftID := old_codec.MustDecodeNFTID([]byte(k))

		// If NFT was added, its data is stored in present in current state.
		// If NFT was deleted, its data is stored in previous state.
		oldOwnerBytes := lo.Ternary(v != nil, v, prevOldNFTToOwner.GetAt(k))
		oldOwner := lo.Must(old_isc.AgentIDFromBytes(oldOwnerBytes))

		newOwner := OldAgentIDtoNewAgentID(oldOwner, oldChainID)
		newObjectRecord := OldNFTIDtoNewObjectRecord(nftID)

		if v != nil {
			w.CreditObjectToAccount(newOwner, *newObjectRecord)
		} else {
			w.DebitObjectFromAccount(newOwner, newObjectRecord.ID)
		}

		count++
		return true
	})

	cli.DebugLogf("Migrated %v NFTs\n", count)
}

func migrateNonce(
	oldState old_kv.KVStoreReader,
	newState kv.KVStore,
	oldChainID old_isc.ChainID,
) {
	cli.DebugLogf("Migrating nonce...\n")

	count := MigrateMapByName(
		oldState,
		newState,
		old_accounts.KeyNonce,
		string(accounts.KeyNonce),
		func(a old_isc.AgentID, nonce *uint64) (isc.AgentID, *uint64) {
			return OldAgentIDtoNewAgentID(a, oldChainID), nonce
		},
	)

	cli.DebugLogf("Migrated %d nonces\n", count)
}
