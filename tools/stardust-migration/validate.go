package main

import (
	"os"
	"strings"

	"github.com/samber/lo"

	"github.com/pmezard/go-difflib/difflib"
	cmd "github.com/urfave/cli/v2"

	old_iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/state/indexedstore"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/tools/stardust-migration/stateaccess/newstate"
	"github.com/iotaledger/wasp/tools/stardust-migration/stateaccess/oldstate"
	"github.com/iotaledger/wasp/tools/stardust-migration/utils/cli"
	"github.com/iotaledger/wasp/tools/stardust-migration/utils/db"
	"github.com/iotaledger/wasp/tools/stardust-migration/validation"
	old_isc "github.com/nnikolash/wasp-types-exported/packages/isc"
	old_kv "github.com/nnikolash/wasp-types-exported/packages/kv"
	old_parameters "github.com/nnikolash/wasp-types-exported/packages/parameters"
	old_state "github.com/nnikolash/wasp-types-exported/packages/state"
	old_indexedstore "github.com/nnikolash/wasp-types-exported/packages/state/indexedstore"
	old_accounts "github.com/nnikolash/wasp-types-exported/packages/vm/core/accounts"
)

func validateMigration(c *cmd.Context) error {
	cli.DebugLoggingEnabled = true

	srcChainDBDir := c.Args().Get(0)
	destChainDBDir := c.Args().Get(1)

	srcKVS := db.ConnectOld(srcChainDBDir)
	srcStore := old_indexedstore.New(old_state.NewStoreWithUniqueWriteMutex(srcKVS))
	oldChainID := old_isc.ChainID(GetAnchorOutput(lo.Must(srcStore.LatestState())).AliasID)

	destKVS := db.ConnectNew(destChainDBDir)
	destStore := indexedstore.New(state.NewStoreWithUniqueWriteMutex(destKVS))
	newChainID := isc.ChainID(GetAnchorObject(lo.Must(destStore.LatestState())).ID)

	// TODO:
	// 1. Check equality of states:
	// 	For each of (latest state and some of intermediate states):
	// 		For each entity (balances, blocklog etc):
	// 			For each entry of that entity:
	// 				Get it using contract call (either API or directly) from old and new node
	// 				What cannot be retrieved using contracts - retrieve directly from db or write additional getters
	// 				Print in a text format
	// 				Compare texts
	// 				Also maybe encode old data using BCS and compare bytes
	//
	// 2. Check results of tracing
	// 	For each of (latest block, some of intermediate blocks, and some manually chosen blocks):
	// 		For each of requests in that block:
	// 			Run tracing of that block on old and new node, and compare results
	// 	As for manually chosen blocks I think about those where GasFeePolicy was changed or where requests failed because of gas.
	//
	// 3. (not sure) Check for unexpected DB keys difference
	// 	For each of (latest state and some of intermediate states):
	// 		Find difference in presence of keys between old and new state
	// 		Filter out keys by prefix, which are expected to not be present
	// 		Analyze the result. Ideally we should eventually have no unexpected difference there
	// 4. Check special accounts and cases: L2Totals, CommonAccount etc.
	// 	We could also do that not over states, but over mutations in blocks.
	// 5. (not sure) Recalculate and validate commitment hash.
	// 6. Specificy check those block, where "rare" objetcs like NFTs are present/changed.

	newLatestState := NewRecordingKVStoreReadOnly(lo.Must(destStore.LatestState()))
	cli.DebugLogf("Latest new state index: %v", newLatestState.R.BlockIndex())

	cli.DebugLogf("Reading old latest state for index #%v...", newLatestState.R.BlockIndex())
	oldLatestState := NewRecordingKVStoreReadOnly(lo.Must(srcStore.StateByIndex(newLatestState.R.BlockIndex())))

	old_parameters.InitL1(&old_parameters.L1Params{
		Protocol: &old_iotago.ProtocolParameters{
			Bech32HRP: old_iotago.PrefixMainnet,
		},
	})

	defer func() {
		if err := recover(); err != nil {
			cli.Logf("Validation paniced")
			PrintLastDBOperations(oldLatestState, newLatestState)
			panic(err)
		}
	}()

	validateStatesEqual(oldLatestState, newLatestState, oldChainID, newChainID)

	return nil
}

func validateStatesEqual(oldState old_kv.KVStoreReader, newState kv.KVStoreReader, oldChainID old_isc.ChainID, newChainID isc.ChainID) {
	cli.DebugLogf("Validating states equality...\n")
	oldStateContentStr := oldStateContentToStr(oldState, oldChainID)
	newStateContentStr := newStateContentToStr(newState, newChainID)

	cli.DebugLogf("Replacing old chain ID with constant placeholer for comparison...")
	oldStateContentStr = strings.Replace(oldStateContentStr, oldChainID.String(), "<chain-id>", -1)

	cli.DebugLogf("Replacing new chain ID with constant placeholer for comparison...")
	newStateContentStr = strings.Replace(newStateContentStr, newChainID.String(), "<chain-id>", -1)

	if oldStateContentStr != newStateContentStr {
		diff, _ := difflib.GetUnifiedDiffString(difflib.UnifiedDiff{
			A:       difflib.SplitLines(oldStateContentStr),
			B:       difflib.SplitLines(newStateContentStr),
			Context: 2,
		})

		cli.DebugLogf("States diff:\n%v\n", diff)
	}

	oldStateFilePath := os.TempDir() + "/stardust-migration-old-state.txt"
	newStateFilePath := os.TempDir() + "/stardust-migration-new-state.txt"
	cli.DebugLogf("Writing old and new states to files %v and %v\n", oldStateFilePath, newStateFilePath)

	os.WriteFile(oldStateFilePath, []byte(oldStateContentStr), 0644)
	os.WriteFile(newStateFilePath, []byte(newStateContentStr), 0644)

	if oldStateContentStr == newStateContentStr {
		cli.DebugLogf("States are equal\n")
	} else {
		cli.DebugLogf("States are NOT equal\n")
		os.Exit(1)
	}
}

func oldStateContentToStr(chainState old_kv.KVStoreReader, chainID old_isc.ChainID) string {
	accountsContractStr := validation.OldAccountsContractContentToStr(oldstate.GetContactStateReader(chainState, old_accounts.Contract.Hname()), chainID)

	return accountsContractStr
}

func newStateContentToStr(chainState kv.KVStoreReader, chainID isc.ChainID) string {
	accountsContractStr := validation.NewAccountsContractContentToStr(newstate.GetContactStateReader(chainState, accounts.Contract.Hname()), chainID)

	return accountsContractStr
}
