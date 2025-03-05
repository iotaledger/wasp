package validation

import (
	"fmt"
	"strings"

	"github.com/samber/lo"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/migrations/allmigrations"
	"github.com/iotaledger/wasp/tools/stardust-migration/utils"
	"github.com/iotaledger/wasp/tools/stardust-migration/utils/cli"
	old_isc "github.com/nnikolash/wasp-types-exported/packages/isc"
	old_kv "github.com/nnikolash/wasp-types-exported/packages/kv"
	old_accounts "github.com/nnikolash/wasp-types-exported/packages/vm/core/accounts"
)

const (
	newSchema = allmigrations.SchemaVersionIotaRebased
)

func OldAccountsContractContentToStr(contractState old_kv.KVStoreReader, chainID old_isc.ChainID) string {
	accsStr, _ := oldAccountsListToStr(contractState, chainID)
	accsLines := strings.Split(accsStr, "\n")
	cli.DebugLogf("Old accounts preview:\n%v\n...", strings.Join(lo.Slice(accsLines, 0, 10), "\n"))

	return accsStr
}

func NewAccountsContractContentToStr(contractState kv.KVStoreReader, chainID isc.ChainID) string {
	accsStr, _ := newAccountsListToStr(contractState, chainID)
	accsLines := strings.Split(accsStr, "\n")
	cli.DebugLogf("New accounts preview:\n%v\n...", strings.Join(lo.Slice(accsLines, 0, 10), "\n"))

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

	cli.DebugLogf("Replacing chain ID with constant placeholder...\n")
	res := fmt.Sprintf("Found %v accounts:\n%v", accs.Len(), utils.SortLines(accsStr.String()))
	res = strings.Replace(res, chainID.String(), "<chain-id>", -1)

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

	cli.DebugLogf("Replacing chain ID with constant placeholder...\n")
	res := fmt.Sprintf("Found %v accounts:\n%v", len(accs), utils.SortLines(accsStr.String()))
	res = strings.Replace(res, chainID.String(), "<chain-id>", -1)

	return res, agentIDs
}

// func oldBaseTokenBalancesToStr(contractState old_kv.KVStoreReader, chainID old_isc.ChainID) string {
// 	cli.DebugLogf("Reading old base token balances...\n")
// 	printProgress, clearProgress := cli.NewProgressPrinter("balances", 0)
// 	defer clearProgress()
// 	var balancesStr strings.Builder
// 	count := 0
// 	// NOTE: Specifically using here prefix iteration instead of using list of accounts.
// 	//       This is done to perform validation using separate logic from the migration - this improved reliability of the validation.
// 	contractState.Iterate(old_accounts.PrefixBaseTokens, func(balanceKey old_kv.Key, v []byte) bool {
// 		accKey := utils.MustRemovePrefix(balanceKey, old_accounts.PrefixBaseTokens)
// 		accID := lo.Must(old_accounts.AgentIDFromKey(old_kv.Key(accKey), chainID))
// 		// NOTE: Using other logic from the one used in migration to improve validation quality.
// 		balance := old_codec.MustDecodeBigIntAbs(v, big.NewInt(0))
// 		balancesStr.WriteString("\t")
// 		balancesStr.WriteString(oldAgentIDToStr(accID))
// 		balancesStr.WriteString(": ")
// 		balancesStr.WriteString(balance.String())
// 		balancesStr.WriteString("\n")
// 		printProgress()
// 		count++

// 		return true
// 	})

// 	cli.DebugLogf("Found %v old base token balances\n", count)
// 	cli.DebugLogf("Replacing chain ID with constant placeholder...\n")
// 	res := fmt.Sprintf("Found %v base token balances:\n%v", count, utils.SortLines(balancesStr.String()))
// 	res = strings.Replace(res, chainID.String(), "<chain-id>", -1)

// 	return res
// }

// func newBaseTokenBalancesToStr(contractState kv.KVStoreReader, chainID isc.ChainID) string {
// 	cli.DebugLogf("Reading new base token balances...\n")
// 	printProgress, clearProgress := cli.NewProgressPrinter("balances", 0)
// 	defer clearProgress()
// 	var balancesStr strings.Builder
// 	count := 0
// 	// NOTE: Specifically using here prefix iteration instead of using list of accounts.
// 	//       This is done to perform validation using separate logic from the migration - this improved reliability of the validation.
// 	contractState.Iterate(accounts.PrefixBaseTokens, func(balanceKey old_kv.Key, v []byte) bool {
// 		accKey := mustRemovePrefix(balanceKey, accounts.PrefixBaseTokens)
// 		accID := lo.Must(old_accounts.AgentIDFromKey(old_kv.Key(accKey), chainID))
// 		// NOTE: Using other logic from the one used in migration to improve validation quality.
// 		balance := old_codec.MustDecodeBigIntAbs(v, big.NewInt(0))
// 		balancesStr.WriteString("\t")
// 		balancesStr.WriteString(oldAgentIDToStr(accID))
// 		balancesStr.WriteString(": ")
// 		balancesStr.WriteString(balance.String())
// 		balancesStr.WriteString("\n")
// 		printProgress()
// 		count++

// 		return true
// 	})

// 	cli.DebugLogf("Found %v new base token balances\n", count)
// 	cli.DebugLogf("Replacing chain ID with constant placeholder...\n")
// 	res := fmt.Sprintf("Found %v base token balances:\n%v", count, utils.SortLines(balancesStr.String()))
// 	res = strings.Replace(res, chainID.String(), "<chain-id>", -1)

// 	return res
// }
