package validation

import (
	"fmt"
	"math/big"
	"strings"

	"github.com/samber/lo"

	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/migrations/allmigrations"
	"github.com/iotaledger/wasp/tools/stardust-migration/utils"
	"github.com/iotaledger/wasp/tools/stardust-migration/utils/cli"
	old_isc "github.com/nnikolash/wasp-types-exported/packages/isc"
	old_kv "github.com/nnikolash/wasp-types-exported/packages/kv"
	old_codec "github.com/nnikolash/wasp-types-exported/packages/kv/codec"
	old_accounts "github.com/nnikolash/wasp-types-exported/packages/vm/core/accounts"
)

const (
	newSchema = allmigrations.SchemaVersionIotaRebased
)

func OldAccountsContractContentToStr(contractState old_kv.KVStoreReader, chainID old_isc.ChainID) string {
	accsStr, _ := oldAccountsListToStr(contractState, chainID)
	cli.DebugLogf("Old accounts preview:\n%v\n", utils.MultilinePreview(accsStr))

	baseTokenBalancesStr := oldBaseTokenBalancesToStr(contractState, chainID)
	cli.DebugLogf("Old base token balances preview:\n%v\n", utils.MultilinePreview(baseTokenBalancesStr))

	return accsStr
}

func NewAccountsContractContentToStr(contractState kv.KVStoreReader, chainID isc.ChainID) string {
	accsStr, _ := newAccountsListToStr(contractState, chainID)
	cli.DebugLogf("New accounts preview:\n%v\n", utils.MultilinePreview(accsStr))

	baseTokenBalancesStr := newBaseTokenBalancesToStr(contractState, chainID)
	cli.DebugLogf("New base token balances preview:\n%v\n", utils.MultilinePreview(baseTokenBalancesStr))

	return accsStr
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
		accsStr.WriteString("\t")
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
		accID := lo.Must(accounts.AgentIDFromKey(kv.Key(accKey), chainID))
		accsStr.WriteString("\t")
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

func oldBaseTokenBalancesToStr(contractState old_kv.KVStoreReader, chainID old_isc.ChainID) string {
	cli.DebugLogf("Reading old base token balances...\n")
	printProgress, clearProgress := cli.NewProgressPrinter("balances", 0)
	defer clearProgress()
	var balancesStr strings.Builder
	count := 0
	// NOTE: Specifically using here prefix iteration instead of using list of accounts.
	//       This is done to perform validation using separate logic from the migration - this improved reliability of the validation.
	contractState.Iterate(old_accounts.PrefixBaseTokens, func(balanceKey old_kv.Key, v []byte) bool {
		accKey := utils.MustRemovePrefix(balanceKey, old_accounts.PrefixBaseTokens)

		var accStr string
		if strings.HasPrefix(string(accKey), old_accounts.L2TotalsAccount) {
			accStr = "L2TotalsAccount"
		} else {
			agentID := lo.Must(old_accounts.AgentIDFromKey(old_kv.Key(accKey), chainID))
			accStr = oldAgentIDToStr(agentID)
		}
		// NOTE: Using other logic from the one used in migration to improve validation quality.
		balance := old_codec.MustDecodeBigIntAbs(v, big.NewInt(0))
		balancesStr.WriteString("\t")
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

func newBaseTokenBalancesToStr(contractState kv.KVStoreReader, chainID isc.ChainID) string {
	cli.DebugLogf("Reading new base token balances...\n")
	printProgress, clearProgress := cli.NewProgressPrinter("balances", 0)
	defer clearProgress()
	var balancesStr strings.Builder
	count := 0
	// NOTE: Specifically using here prefix iteration instead of using list of accounts.
	//       This is done to perform validation using separate logic from the migration - this improved reliability of the validation.
	contractState.Iterate(kv.Key(accounts.PrefixAccountCoinBalances), func(balanceKey kv.Key, v []byte) bool {
		accKey, coinTypeBytes := utils.SplitMapKey(balanceKey, accounts.PrefixAccountCoinBalances)

		if coinTypeBytes == "" {
			// not a map entry
			return true
		}

		coinType := lo.Must(coin.TypeFromBytes([]byte(coinTypeBytes)))
		if coinType != coin.BaseTokenType {
			// Native token - ignoring
			return true
		}

		var accStr string
		if accKey == accounts.L2TotalsAccount {
			accStr = "L2TotalsAccount"
		} else {
			agentID := lo.Must(accounts.AgentIDFromKey(kv.Key(accKey), chainID))
			accStr = newAgentIDToStr(agentID)
		}

		balance := codec.MustDecode[coin.Value](v)
		balanceFullDecimal := util.BaseTokensDecimalsToEthereumDecimals(balance, parameters.Decimals)

		var remeinder *big.Int
		if remeinderBytes := contractState.Get(accounts.AccountWeiRemainderKey(accKey)); remeinderBytes != nil {
			remeinder = codec.MustDecode[*big.Int](contractState.Get(accounts.AccountWeiRemainderKey(accKey)))
			balanceFullDecimal.Add(balanceFullDecimal, remeinder)
		}

		// NOTE: Using other logic from the one used in migration to improve validation quality.
		balancesStr.WriteString("\t")
		balancesStr.WriteString(accStr)
		balancesStr.WriteString(": ")
		balancesStr.WriteString(balanceFullDecimal.String())
		balancesStr.WriteString("\n")
		printProgress()
		count++

		return true
	})

	cli.DebugLogf("Found %v new base token balances\n", count)
	cli.DebugLogf("Formatting lines...\n")
	res := fmt.Sprintf("Found %v base token balances:%v", count, utils.SortLines(balancesStr.String()))

	return res
}
