package validation

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/samber/lo"

	old_isc "github.com/nnikolash/wasp-types-exported/packages/isc"
	old_kv "github.com/nnikolash/wasp-types-exported/packages/kv"
	old_codec "github.com/nnikolash/wasp-types-exported/packages/kv/codec"
	old_accounts "github.com/nnikolash/wasp-types-exported/packages/vm/core/accounts"

	old_iotago "github.com/iotaledger/iota.go/v3"

	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/tools/stardust-migration/utils"
	"github.com/iotaledger/wasp/tools/stardust-migration/utils/cli"
)

func OldAccountsContractContentToStr(contractState old_kv.KVStoreReader, chainID old_isc.ChainID) string {
	// TODO: There will be not enought memory to store all that stringyfied data.
	// Would need to change the flow of validation. But for current development it's fine.

	accsStr, accs := oldAccountsListToStr(contractState, chainID)
	cli.DebugLogf("Old accounts preview:\n%v\n", utils.MultilinePreview(accsStr))

	baseTokenBalancesStr := oldBaseTokenBalancesToStr(contractState, chainID, accs)
	cli.DebugLogf("Old base token balances preview:\n%v\n", utils.MultilinePreview(baseTokenBalancesStr))

	nativeTokenBalancesStr := oldNativeTokenBalancesToStr(contractState, chainID, accs)
	cli.DebugLogf("Old native token balances preview:\n%v\n", utils.MultilinePreview(nativeTokenBalancesStr))

	return accsStr + "\n" + baseTokenBalancesStr + "\n" + nativeTokenBalancesStr
}

func NewAccountsContractContentToStr(contractState kv.KVStoreReader, chainID isc.ChainID) string {
	accsStr, accs := newAccountsListToStr(contractState, chainID)
	cli.DebugLogf("New accounts preview:\n%v\n", utils.MultilinePreview(accsStr))

	baseTokenBalancesStr, nativeTOkenBalancesStr := newTokenBalancesToStr(contractState, chainID, accs)
	cli.DebugLogf("New base token balances preview:\n%v\n", utils.MultilinePreview(baseTokenBalancesStr))
	cli.DebugLogf("New native token balances preview:\n%v\n", utils.MultilinePreview(nativeTOkenBalancesStr))

	return accsStr + "\n" + baseTokenBalancesStr + "\n" + nativeTOkenBalancesStr
}

func oldAccountsListToStr(contractState old_kv.KVStoreReader, chainID old_isc.ChainID) (string, map[old_kv.Key]old_isc.AgentID) {
	cli.DebugLogf("Reading old accounts list...\n")
	accs := old_accounts.AllAccountsMapR(contractState)

	cli.DebugLogf("Found %v accounts\n", accs.Len())
	cli.DebugLogf("Reading accounts...\n")
	printProgress, clearProgress := cli.NewProgressPrinter("accounts", accs.Len())
	defer clearProgress()

	var accsStr strings.Builder
	agentIDs := make(map[old_kv.Key]old_isc.AgentID)
	accs.Iterate(func(accKey []byte, accValue []byte) bool {
		accID := lo.Must(old_accounts.AgentIDFromKey(old_kv.Key(accKey), chainID))
		accsStr.WriteString("\tAcc: ")
		accsStr.WriteString(oldAgentIDToStr(accID))
		accsStr.WriteString("\n")
		agentIDs[old_kv.Key(accKey)] = accID
		printProgress()
		return true
	})

	cli.DebugLogf("Formatting lines...\n")
	res := fmt.Sprintf("Found %v accounts:%v", accs.Len(), utils.SortLines(accsStr.String()))

	return res, agentIDs
}

func newAccountsListToStr(contractState kv.KVStoreReader, chainID isc.ChainID) (string, map[kv.Key]isc.AgentID) {
	cli.DebugLogf("Reading new accounts list...\n")
	accs := accounts.NewStateReader(newSchema, contractState).AllAccountsAsDict()

	cli.DebugLogf("Found %v accounts\n", len(accs))
	cli.DebugLogf("Reading accounts...\n")
	printProgress, clearProgress := cli.NewProgressPrinter("accounts", uint32(len(accs)))
	defer clearProgress()

	var accsStr strings.Builder
	agentIDs := make(map[kv.Key]isc.AgentID)
	accs.Iterate("", func(accKey kv.Key, accValue []byte) bool { // NOTE: using Iterate instead of IterateSorted because lines will be sorted anyway
		accID := lo.Must(accounts.AgentIDFromKey(kv.Key(accKey)))
		accsStr.WriteString("\tAcc: ")
		accsStr.WriteString(newAgentIDToStr(accID))
		accsStr.WriteString("\n")
		agentIDs[kv.Key(accKey)] = accID
		printProgress()
		return true
	})

	cli.DebugLogf("Formatting lines...\n")
	res := fmt.Sprintf("Found %v accounts:%v", len(accs), utils.SortLines(accsStr.String()))

	return res, agentIDs
}

func oldBaseTokenBalancesToStr(contractState old_kv.KVStoreReader, chainID old_isc.ChainID, knownAccs map[old_kv.Key]old_isc.AgentID) string {
	balancesStrFromPrefix := oldBaseTokenBalancesFromPrefixToStr(contractState, chainID, knownAccs)
	balancesStrFromMap := oldBaseTokenBalancesFromMapToStr(contractState, chainID, knownAccs)

	EnsureEqual("old base balances (prefix vs map)", balancesStrFromPrefix, balancesStrFromMap)

	return balancesStrFromPrefix
}

func oldBaseTokenBalancesFromPrefixToStr(contractState old_kv.KVStoreReader, chainID old_isc.ChainID, knownAccs map[old_kv.Key]old_isc.AgentID) string {
	cli.DebugLogf("Reading old base token balances (by prefix)...\n")

	printProgress, clearProgress := cli.NewProgressPrinter("balances", 0)
	defer clearProgress()

	var balancesStr strings.Builder
	count := 0

	// NOTE: Specifically using here prefix iteration instead of using list of accounts.
	//       This is done to perform validation using separate logic from the migration - this improved reliability of the validation.
	contractState.Iterate(old_accounts.PrefixBaseTokens, func(k old_kv.Key, v []byte) bool {
		accKey := utils.MustRemovePrefix(k, old_accounts.PrefixBaseTokens)

		var accStr string
		if accKey == old_accounts.L2TotalsAccount {
			accStr = "L2TotalsAccount"
		} else {
			agentID := lo.Must(old_accounts.AgentIDFromKey(old_kv.Key(accKey), chainID))
			accStr = oldAgentIDToStr(agentID)

			knownAgentID, ok := knownAccs[accKey]
			if !ok {
				panic(fmt.Errorf("account has balance, but not found in accounts list: agentID = %v, accKey = %x / %v", accStr, accKey, string(accKey)))
			}

			knownAgentIDStr := oldAgentIDToStr(knownAgentID)
			if knownAgentIDStr != accStr {
				panic(fmt.Errorf("differnt agent ID for same acc key: knownAgentID = %v, balanceAgentID = %v, accKey = %x / %v",
					knownAgentIDStr, accStr, accKey, string(accKey)))
			}
		}

		// NOTE: Using other logic from the one used in migration to improve validation quality.
		balance := old_codec.MustDecodeBigIntAbs(v, big.NewInt(0))
		balancesStr.WriteString("\tBase balance: ")
		balancesStr.WriteString(accStr)
		balancesStr.WriteString(": ")
		balancesStr.WriteString(balance.String())
		balancesStr.WriteString("\n")
		printProgress()
		count++

		return true
	})

	cli.DebugLogf("Found %v old base token balances\n", count)
	cli.DebugLogf("Formatting lines...\n")
	res := fmt.Sprintf("Found %v base token balances:%v", count, utils.SortLines(balancesStr.String()))

	return res
}

func oldBaseTokenBalancesFromMapToStr(contractState old_kv.KVStoreReader, chainID old_isc.ChainID, knownAccs map[old_kv.Key]old_isc.AgentID) string {
	cli.DebugLogf("Reading old base token balances (from map)...\n")

	printProgress, clearProgress := cli.NewProgressPrinter("balances", len(knownAccs)+1)
	defer clearProgress()

	var balancesStr strings.Builder
	count := 0

	stringifyBalance := func(accKey old_kv.Key, accStr string, balance *big.Int) {
		if balance.Sign() == 0 {
			return
		}

		balancesStr.WriteString("\tBase balance: ")
		balancesStr.WriteString(accStr)
		balancesStr.WriteString(": ")
		balancesStr.WriteString(balance.String())
		balancesStr.WriteString("\n")
		printProgress()
		count++
	}

	for accKey, agentID := range knownAccs {
		balance := old_accounts.GetBaseTokensFullDecimals(newSchema)(contractState, accKey)
		accStr := oldAgentIDToStr(agentID)
		stringifyBalance(accKey, accStr, balance)
	}

	totalsBalance := old_accounts.GetBaseTokensFullDecimals(newSchema)(contractState, old_kv.Key(old_accounts.L2TotalsAccount))
	stringifyBalance(old_kv.Key(old_accounts.L2TotalsAccount), "L2TotalsAccount", totalsBalance)

	cli.DebugLogf("Found %v old base token balances\n", count)
	cli.DebugLogf("Formatting lines...\n")
	res := fmt.Sprintf("Found %v base token balances:%v", count, utils.SortLines(balancesStr.String()))

	return res
}

func oldNativeTokenBalancesToStr(contractState old_kv.KVStoreReader, chainID old_isc.ChainID, knownAccs map[old_kv.Key]old_isc.AgentID) string {
	balancesStrFromPrefix := oldNativeTokenBalancesFromPrefixToStr(contractState, chainID, knownAccs)

	return balancesStrFromPrefix
}

func oldNativeTokenBalancesFromPrefixToStr(contractState old_kv.KVStoreReader, chainID old_isc.ChainID, knownAccs map[old_kv.Key]old_isc.AgentID) string {
	cli.DebugLogf("Reading old native token balances (by prefix)...\n")

	printProgress, clearProgress := cli.NewProgressPrinter("balances", 0)
	defer clearProgress()

	var balancesStr strings.Builder
	count := 0

	contractState.Iterate(old_accounts.PrefixNativeTokens, func(k old_kv.Key, v []byte) bool {
		accKey, accStr, _, ntID, isMapElem := utils.MustSplitParseMapKeyAny(k, old_accounts.PrefixNativeTokens, func(accKey, ntIDBytes old_kv.Key) (string, old_iotago.NativeTokenID, error) {
			var accStr string
			if accKey == old_accounts.L2TotalsAccount {
				accStr = "L2TotalsAccount"
			} else {
				agentID, err := old_accounts.AgentIDFromKey(old_kv.Key(accKey), chainID)
				if err != nil {
					return "", old_iotago.NativeTokenID{}, fmt.Errorf("failed to parse agent ID: %v", err)
				}

				accStr = oldAgentIDToStr(agentID)
			}

			ntID, err := old_isc.NativeTokenIDFromBytes([]byte(ntIDBytes))
			if err != nil {
				return "", old_iotago.NativeTokenID{}, fmt.Errorf("failed to parse native token ID: %v: ntIDBytes = %x / %v", err, ntIDBytes, string(ntIDBytes))
			}

			return accStr, ntID, nil
		})
		if !isMapElem {
			return true
		}

		if accKey != old_accounts.L2TotalsAccount {
			knownAgentID, ok := knownAccs[accKey]
			if !ok {
				panic(fmt.Errorf("account has balance, but not found in accounts list: agentID = %v, accKey = %x / %v", accStr, accKey, string(accKey)))
			}

			knownAgentIDStr := oldAgentIDToStr(knownAgentID)
			if knownAgentIDStr != accStr {
				panic(fmt.Errorf("differnt agent ID for same acc key: knownAgentID = %v, balanceAgentID = %v, accKey = %x / %v",
					knownAgentIDStr, accStr, accKey, string(accKey)))
			}
		}

		// NOTE: Using other logic from the one used in migration to improve validation quality.
		balance := old_codec.MustDecodeBigIntAbs(v, big.NewInt(0))
		balancesStr.WriteString("\tNative balance: ")
		balancesStr.WriteString(accStr)
		balancesStr.WriteString(": ")
		balancesStr.WriteString(ntID.String())
		balancesStr.WriteString(": ")
		balancesStr.WriteString(balance.String())
		balancesStr.WriteString("\n")
		printProgress()
		count++

		return true
	})

	cli.DebugLogf("Found %v old native token balances\n", count)
	cli.DebugLogf("Formatting lines...\n")
	res := fmt.Sprintf("Found %v native token balances:%v", count, utils.SortLines(balancesStr.String()))

	return res
}

func newTokenBalancesToStr(contractState kv.KVStoreReader, chainID isc.ChainID, accs map[kv.Key]isc.AgentID) (base, native string) {
	// Using two different ways of getting balances and ensuring they are equal - for double safety
	baseFromPrefix, nativeFromPrefix := newTokenBalancesFromPrefixToStr(contractState, chainID)
	baseFromMap, nativeFromMap := newTokenBalancesFromMapToStr(contractState, chainID, accs)

	EnsureEqual("new base token balances (prefix vs map)", baseFromPrefix, baseFromMap)
	EnsureEqual("new native token balances (prefix vs map)", nativeFromPrefix, nativeFromMap)

	return baseFromPrefix, nativeFromPrefix
}

func newTokenBalancesFromPrefixToStr(contractState kv.KVStoreReader, chainID isc.ChainID) (base, native string) {
	cli.DebugLogf("Reading new token balances (using prefix iteration)...\n")

	printProgress, clearProgress := cli.NewProgressPrinter("balances", 0)
	defer clearProgress()

	var baseBalancesStr strings.Builder
	var nativeBalancesStr strings.Builder
	baseCount := 0
	nativeCount := 0

	// NOTE: Specifically using here prefix iteration instead of using list of accounts.
	//       This is done to perform validation using separate logic from the migration - this improved reliability of the validation.
	contractState.Iterate(kv.Key(accounts.PrefixAccountCoinBalances), func(k kv.Key, v []byte) bool {
		accKey, accStr, _, coinType, isMapElem := utils.MustSplitParseMapKeyAny(k, accounts.PrefixAccountCoinBalances, func(accKey, coinTypeBytes kv.Key) (string, coin.Type, error) {
			// Unfortunatelly sometimes accKey or coinTypeBytes contains map separator (dot - .)
			// And as both accKey and coinTypeBytes hae dynamic size, we cannot expected the separator at some specific position.
			// So what we do is just try to parse all variants.

			var accStr string
			if accKey == accounts.L2TotalsAccount {
				accStr = "L2TotalsAccount"
			} else {
				agentID, err := accounts.AgentIDFromKey(kv.Key(accKey))
				if err != nil {
					return "", coin.Type{}, fmt.Errorf("failed to parse agent ID: %v", err)
				}
				accStr = newAgentIDToStr(agentID)
			}

			coinType, err := coin.TypeFromBytes([]byte(coinTypeBytes))
			if err != nil {
				return "", coin.Type{}, fmt.Errorf("failed to parse coin type: %v: coinTypeBytes = %x / %v", err, coinTypeBytes, string(coinTypeBytes))
			}

			return accStr, coinType, nil
		})
		if !isMapElem {
			return true
		}

		balance := codec.MustDecode[coin.Value](v)

		var balanceStr string
		var strBuilder *strings.Builder
		if coinType == coin.BaseTokenType {
			balanceFullDecimal := util.BaseTokensDecimalsToEthereumDecimals(balance, parameters.BaseTokenDecimals)

			var remeinder *big.Int
			if remeinderBytes := contractState.Get(accounts.AccountWeiRemainderKey(accKey)); remeinderBytes != nil {
				remeinder = codec.MustDecode[*big.Int](contractState.Get(accounts.AccountWeiRemainderKey(accKey)))
				balanceFullDecimal.Add(balanceFullDecimal, remeinder)
			}

			// Do not need to convert anythng - full decimal form stayed same.

			balanceStr = balanceFullDecimal.String()
			strBuilder = &baseBalancesStr
			baseCount++
			strBuilder.WriteString("\tBase balance: ")
		} else {
			ntID := coinTypeToOldNTID(coinType) // reverse conversion
			balanceStr = ntID.ToHex() + ": " + balance.String()
			strBuilder = &nativeBalancesStr
			nativeCount++
			strBuilder.WriteString("\tNative balance: ")
		}

		strBuilder.WriteString(accStr)
		strBuilder.WriteString(": ")
		strBuilder.WriteString(balanceStr)
		strBuilder.WriteString("\n")

		printProgress()

		return true
	})

	// Process balances with remainder but without coin balance part
	contractState.Iterate(accounts.PrefixAccountWeiRemainder, func(k kv.Key, v []byte) bool {
		accKey := utils.MustRemovePrefix(k, accounts.PrefixAccountWeiRemainder)
		coinBalance := accounts.NewStateReader(newSchema, contractState).UnsafeGetCoinBalance(accKey, coin.BaseTokenType)
		if coinBalance != 0 {
			return true
		}

		agentID := lo.Must(accounts.AgentIDFromKey(accKey))
		remainder := codec.MustDecode[*big.Int](v)

		balanceFullDecimal := util.BaseTokensDecimalsToEthereumDecimals(0, parameters.BaseTokenDecimals)
		balanceFullDecimal.Add(balanceFullDecimal, remainder)

		baseBalancesStr.WriteString("\tBase balance: ")
		baseBalancesStr.WriteString(newAgentIDToStr(agentID))
		baseBalancesStr.WriteString(": ")
		baseBalancesStr.WriteString(balanceFullDecimal.String())
		baseBalancesStr.WriteString("\n")

		baseCount++

		return true
	})

	cli.DebugLogf("Found %v new base token balances, %v new native token balances\n", baseCount, nativeCount)
	cli.DebugLogf("Formatting lines...\n")
	resBase := fmt.Sprintf("Found %v base token balances:%v", baseCount, utils.SortLines(baseBalancesStr.String()))
	resNative := fmt.Sprintf("Found %v native token balances:%v", nativeCount, utils.SortLines(nativeBalancesStr.String()))

	return resBase, resNative
}

func newTokenBalancesFromMapToStr(contractState kv.KVStoreReader, chainID isc.ChainID, accs map[kv.Key]isc.AgentID) (base, native string) {
	cli.DebugLogf("Reading new token balances (using accs map)...\n")

	printProgress, clearProgress := cli.NewProgressPrinter("balances", uint32(len(accs)))
	defer clearProgress()

	var baseBalancesStr strings.Builder
	var nativeBalancesStr strings.Builder
	baseCount := 0
	nativeCount := 0

	addBalanceStr := func(accKey kv.Key, agentIDStr string, balanceStr string, coinType coin.Type) {
		defer printProgress()

		if balanceStr == "0" {
			return
		}

		var strBuilder *strings.Builder

		if coinType == coin.BaseTokenType {
			strBuilder = &baseBalancesStr
			baseCount++
			strBuilder.WriteString("\tBase balance: ")
		} else {
			strBuilder = &nativeBalancesStr
			ntID := coinTypeToOldNTID(coinType) // reverse conversion
			balanceStr = ntID.ToHex() + ": " + balanceStr
			nativeCount++
			strBuilder.WriteString("\tNative balance: ")
		}

		strBuilder.WriteString(agentIDStr)
		strBuilder.WriteString(": ")
		strBuilder.WriteString(balanceStr)
		strBuilder.WriteString("\n")
	}

	r := accounts.NewStateReader(newSchema, contractState)

	for accKey, agentID := range accs {
		baseBalance := r.GetBaseTokensBalanceFullDecimals(agentID)
		addBalanceStr(accKey, newAgentIDToStr(agentID), baseBalance.String(), coin.BaseTokenType)

		nativeTokens := r.GetAccountFungibleTokens(agentID).NativeTokens()
		for coinType, ntBalance := range nativeTokens {
			if coinType == coin.BaseTokenType {
				continue
			}
			addBalanceStr(accKey, newAgentIDToStr(agentID), ntBalance.String(), coinType)
		}
	}

	totalBaseTokens := r.UnsafeGetBaseTokensFullDecimals(accounts.L2TotalsAccount)
	addBalanceStr(accounts.L2TotalsAccount, "L2TotalsAccount", totalBaseTokens.String(), coin.BaseTokenType)

	totalNativeTokens := r.GetTotalL2FungibleTokens()
	for coinType, ntBalance := range totalNativeTokens {
		if coinType == coin.BaseTokenType {
			continue
		}
		addBalanceStr(accounts.L2TotalsAccount, "L2TotalsAccount", ntBalance.String(), coinType)
	}

	cli.DebugLogf("Found %v new base token balances, %v new native token balances\n", baseCount, nativeCount)
	cli.DebugLogf("Formatting lines...\n")
	resBase := fmt.Sprintf("Found %v base token balances:%v", baseCount, utils.SortLines(baseBalancesStr.String()))
	resNative := fmt.Sprintf("Found %v native token balances:%v", nativeCount, utils.SortLines(nativeBalancesStr.String()))

	return resBase, resNative
}

func coinTypeToOldNTID(t coin.Type) old_iotago.NativeTokenID {
	rt := t.ResourceType()
	if rt.Module != "nt" || !strings.HasPrefix(rt.ObjectName, "NT") {
		// Yes, raw comparison with magic strings. Intended - to make validation less dependant on bugs of the migration code.
		panic(fmt.Errorf("unexpected native token type: %v: %v, %v", t, rt.Module, rt.ObjectName))
	}

	addr := rt.Address.Bytes()
	foundrySerialNo := lo.Must(hexutil.Decode("0x" + strings.TrimPrefix(rt.ObjectName, "NT"))) // will be wrong if zero-padded
	return old_iotago.NativeTokenID(append(addr, foundrySerialNo...))
}
