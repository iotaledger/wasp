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

func MigrateAccountsContractMuts(
	v old_isc.SchemaVersion,
	oldChainStateMuts old_kv.KVStoreReader,
	newChainState kv.KVStore,
	oldChainID old_isc.ChainID,
) {
	cli.DebugLog("Migrating accounts contract (muts)...\n")
	oldStateMuts := oldstate.GetContactStateReader(oldChainStateMuts, old_accounts.Contract.Hname())
	newState := newstate.GetContactState(newChainState, accounts.Contract.Hname())

	migrateAccountsList(oldStateMuts, newState, oldChainID)
	migrateBaseTokenBalances(v, oldStateMuts, newState, oldChainID)
	migrateNativeTokenBalances(oldStateMuts, newState, oldChainID)
	migrateFoundriesOutputs(oldStateMuts, newState)
	//migrateFoundriesPerAccount(oldState, newState, oldChainID)
	migrateNativeTokenOutputs(oldStateMuts, newState)
	// migrateNFTs is done in MigrateAccountsContractFullState
	// prefixNewlyMintedNFTs ignored
	// PrefixMintIDMap ignored
	migrateNonce(oldStateMuts, newState, oldChainID)

	cli.DebugLog("Migrated accounts contract (muts)\n")
}

func MigrateAccountsContractFullState(
	oldChainState old_kv.KVStoreReader,
	newChainState kv.KVStore,
	oldChainID old_isc.ChainID,
) {
	// NOTE: See comment in migrateNFTs for reason why MigrateAccountsContract is split into two functions.
	cli.DebugLog("Migrating accounts contract (full state)...\n")
	oldState := oldstate.GetContactStateReader(oldChainState, old_accounts.Contract.Hname())
	newState := newstate.GetContactState(newChainState, accounts.Contract.Hname())

	migrateNFTs(oldState, newState, oldChainID)

	cli.DebugLog("Migrated accounts contract (full state)\n")
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
			if !IsValidOldAccountKeyBytesLen(len(oldAccKey)) {
				// TODO: what is that key on block 135355 with length 7?
				//       oldAccKey = 6b79a90c08c1b
				//		 k = 74e6b79a90c08c1b2edc0c0dd4357c3d8db0d7302f
				return true
			}

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
	oldChainID old_isc.ChainID,
	newChainID isc.ChainID,
) {
	cli.DebugLogf("Migrating foundries of accounts...\n")

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
	oldState old_kv.KVStoreReader,
	newState kv.KVStore,
	oldChainID old_isc.ChainID,
) {
	// NOTE: We cant use just mutations for this function, because GetNFTData() reads other stuff from the state.
	// And even if we pass here both mutations and full state and iterate by mutations but read from full state,
	// still it wouldn't work, because old state with applied mutations will not have the data read by GetNFTData().
	// So to use mutations here, we need to have both mutations and PREV old state. That seems too complicated for now,
	// just using full state here as special case of accounts contract migration.

	cli.DebugLogf("Migrating NFTs...\n")

	oldNFTOutputs := old_accounts.NftOutputMapR(oldState)
	w := accounts.NewStateWriter(newSchema, newState)

	var count uint32
	oldNFTOutputs.IterateKeys(func(k []byte) bool {
		nftID := old_codec.MustDecodeNFTID([]byte(k))
		oldNFT := old_accounts.GetNFTData(oldState, nftID)
		owner := OldAgentIDtoNewAgentID(oldNFT.Owner, oldChainID)
		newObjectRecord := OldNFTIDtoNewObjectRecord(nftID)
		w.CreditObjectToAccount(owner, *newObjectRecord)
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
