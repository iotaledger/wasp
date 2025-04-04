package main

import (
	"os"
	"strings"

	"github.com/samber/lo"

	cmd "github.com/urfave/cli/v2"

	old_isc "github.com/nnikolash/wasp-types-exported/packages/isc"
	old_kv "github.com/nnikolash/wasp-types-exported/packages/kv"
	old_parameters "github.com/nnikolash/wasp-types-exported/packages/parameters"
	old_state "github.com/nnikolash/wasp-types-exported/packages/state"
	old_indexedstore "github.com/nnikolash/wasp-types-exported/packages/state/indexedstore"
	old_accounts "github.com/nnikolash/wasp-types-exported/packages/vm/core/accounts"
	old_blocklog "github.com/nnikolash/wasp-types-exported/packages/vm/core/blocklog"

	old_iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/state/indexedstore"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/tools/stardust-migration/stateaccess/newstate"
	"github.com/iotaledger/wasp/tools/stardust-migration/stateaccess/oldstate"
	"github.com/iotaledger/wasp/tools/stardust-migration/utils/cli"
	"github.com/iotaledger/wasp/tools/stardust-migration/utils/db"
	"github.com/iotaledger/wasp/tools/stardust-migration/validation"
)

func validateMigration(c *cmd.Context) error {
	go func() {
		<-c.Done()
		cli.Logf("Interrupted")
		os.Exit(1)
	}()

	firstIndex := uint32(c.Uint64("from-block"))
	validation.ConcurrentValidation = !c.Bool("no-parallel")
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
	// 3. Check special accounts and cases: L2Totals, CommonAccount etc.
	// 4. Specifically check those block, where "rare" objects like NFTs are present/change
	// 5. Specifically check at least one instance of EACH business entity (e.g. OnLedgerRequest, OffLedgerRequest, etc.)
	// 6. Into each of new state validations add some basic "integration" tests - use business logic funcs to retrieve data
	// 7. Perform ALL callviews at least once for each of business entity type

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
			cli.Logf("Validation panicked")
			PrintLastDBOperations(oldLatestState, newLatestState)
			panic(err)
		}
	}()

	validateStatesEqual(oldLatestState, newLatestState, oldChainID, newChainID, firstIndex, newLatestState.R.BlockIndex())

	return nil
}

func validateStatesEqual(oldState old_kv.KVStoreReader, newState kv.KVStoreReader, oldChainID old_isc.ChainID, newChainID isc.ChainID, firstIndex, lastIndex uint32) {
	cli.DebugLogf("Validating states equality...\n")
	var oldStateContentStr, newStateContentStr string
	validation.GoAllAndWait(func() {
		oldStateContentStr = oldStateContentToStr(oldState, oldChainID, firstIndex, lastIndex)
		cli.DebugLogf("Replacing old chain ID with constant placeholer for comparison...")
		oldStateContentStr = strings.Replace(oldStateContentStr, oldChainID.String(), "<chain-id>", -1)

	}, func() {
		newStateContentStr = newStateContentToStr(newState, newChainID, firstIndex, lastIndex)
		cli.DebugLogf("Replacing new chain ID with constant placeholer for comparison...")
		newStateContentStr = strings.Replace(newStateContentStr, newChainID.String(), "<chain-id>", -1)
	})

	validation.EnsureEqual("states", oldStateContentStr, newStateContentStr)

	cli.DebugLogf("States are equal\n")
}

func oldStateContentToStr(chainState old_kv.KVStoreReader, chainID old_isc.ChainID, firstIndex, lastIndex uint32) string {
	var accountsContractStr, blocklogContractStr string

	validation.GoAllAndWait(func() {
		accountsContractStr = validation.OldAccountsContractContentToStr(oldstate.GetContactStateReader(chainState, old_accounts.Contract.Hname()), chainID)
	}, func() {
		blocklogContractStr = validation.OldBlocklogContractContentToStr(oldstate.GetContactStateReader(chainState, old_blocklog.Contract.Hname()), chainID, firstIndex, lastIndex)
	})

	return accountsContractStr + "\n" + blocklogContractStr
}

func newStateContentToStr(chainState kv.KVStoreReader, chainID isc.ChainID, firstIndex, lastIndex uint32) string {
	var accountsContractStr, blocklogContractStr string
	validation.GoAllAndWait(func() {
		accountsContractStr = validation.NewAccountsContractContentToStr(newstate.GetContactStateReader(chainState, accounts.Contract.Hname()), chainID)
	}, func() {
		blocklogContractStr = validation.NewBlocklogContractContentToStr(newstate.GetContactStateReader(chainState, blocklog.Contract.Hname()), chainID, firstIndex, lastIndex)
	})

	return accountsContractStr + "\n" + blocklogContractStr
}
