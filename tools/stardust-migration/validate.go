package main

import (
	"fmt"
	"math"
	"os"
	"runtime/debug"
	"slices"
	"strconv"
	"strings"

	"github.com/samber/lo"

	cmd "github.com/urfave/cli/v2"

	old_isc "github.com/nnikolash/wasp-types-exported/packages/isc"
	old_kv "github.com/nnikolash/wasp-types-exported/packages/kv"
	old_parameters "github.com/nnikolash/wasp-types-exported/packages/parameters"
	old_state "github.com/nnikolash/wasp-types-exported/packages/state"
	old_indexedstore "github.com/nnikolash/wasp-types-exported/packages/state/indexedstore"

	old_iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/state/indexedstore"
	"github.com/iotaledger/wasp/tools/stardust-migration/utils"
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
	lastIndex := c.Uint64("to-block")
	short := c.Bool("short")
	blocksListFile := c.String("blocks-list")
	validation.ConcurrentValidation = !c.Bool("no-parallel")
	hashValues := !c.Bool("no-hashing")
	findFailureBlock := c.Bool("find-fail-block")
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
	// 			* Run tracing of that block on old and new node, and compare results
	//          * Check that tracing of all requests in the block produce same mutation as in the block
	//          * Check that mutations are equal in terms of order of elements in RequestLookupKeysList
	// 	As for manually chosen blocks I think about those where
	//   * GasFeePolicy was changed or where requests failed because of gas.
	//   * Before and after pruning config changed
	//
	// 3. Check special accounts and cases: L2Totals, CommonAccount etc.
	// 4. Specifically check those block, where "rare" objects like NFTs are present/change
	// 5. Specifically check at least one instance of EACH business entity (e.g. OnLedgerRequest, OffLedgerRequest, etc.)
	// 6. Check prunning is duplicated: requests, lookup table, and others are not available when pruned in old state.
	// 7. Into each of new state validations add some basic "integration" tests - use business logic funcs to retrieve data.
	// 8. Perform ALL callviews at least once for each of business entity type.
	// 9. Ensure hash of EVM transaction tracing is same for transactions with Native Tokens and NFTs
	// 10. Validate that commitments does not change between multiple migration runs

	if lastIndex == math.MaxUint64 {
		cli.Logf("Using latest new state index")
		lastIndex = uint64(lo.Must(destStore.LatestBlockIndex()))
	}

	cli.Logf("State index to be validated: %v", lastIndex)

	old_parameters.InitL1(&old_parameters.L1Params{
		Protocol: &old_iotago.ProtocolParameters{
			Bech32HRP: old_iotago.PrefixMainnet,
		},
	})

	var lastErr error

	validation.HashValues = hashValues
	switch {
	case !findFailureBlock && blocksListFile == "":
		cli.DebugLoggingEnabled = true

		cli.Logf("Reading old latest state for index #%v...", lastIndex)
		oldState := utils.NewRecordingKVStoreReadOnly(lo.Must(srcStore.StateByIndex(uint32(lastIndex))))

		cli.Logf("Reading new latest state for index #%v...", lastIndex)
		newState := utils.NewRecordingKVStoreReadOnly(lo.Must(destStore.StateByIndex(uint32(lastIndex))))

		defer func() {
			if err := recover(); err != nil {
				cli.Logf("Validation panicked")
				utils.PrintLastDBOperations(oldState, newState)
				panic(err)
			}
		}()

		validateStatesEqual(oldState, newState, oldChainID, newChainID, firstIndex, uint32(lastIndex), short)
		return nil
	case findFailureBlock && blocksListFile == "":
		cli.DebugLoggingEnabled = false

		findBlockWithIssue(firstIndex, uint32(lastIndex), func(blockIndex uint32) bool {
			cli.Logf("Reading old latest state for index #%v...", blockIndex)
			oldState := utils.NewRecordingKVStoreReadOnly(lo.Must(srcStore.StateByIndex(blockIndex)))

			cli.Logf("Reading new latest state for index #%v...", blockIndex)
			newState := utils.NewRecordingKVStoreReadOnly(lo.Must(destStore.StateByIndex(blockIndex)))

			defer func() {
				if err := recover(); err != nil {
					cli.Logf("Validation panicked")
					lastErr = fmt.Errorf("%v\n%v", err, string(debug.Stack()))
				}
			}()

			validateStatesEqual(oldState, newState, oldChainID, newChainID, firstIndex, blockIndex, short)
			return true
		})

		if lastErr != nil {
			cli.ClearStatusBar()
			cli.Logf("Validation panicked: %v", lastErr)
		}
	case blocksListFile != "" && !findFailureBlock:
		// Could be done in parallel, but it may use too much ram, so not risking.
		cli.DebugLoggingEnabled = false
		validation.ProgressEnabled = false

		cli.Logf("Reading blocks list from %s...", blocksListFile)
		blocks := lo.Must(readBlocksList(blocksListFile))
		failedBlocks := make([]uint32, 0)

		cli.Logf("Validating %v blocks", len(blocks))

		i := 0
		if firstIndex != 0 {
			for len(blocks) != 0 {
				if blocks[i] >= firstIndex {
					break
				}
				cli.Logf("Skipping block %v (%v/%v) because it is before the first index %v", blocks[i], i+1, len(blocks), firstIndex)
				i++
			}

			if len(blocks) == 0 {
				cli.Logf("Start block %v not found in the list", firstIndex)
				return nil
			}
			if blocks[i] != firstIndex {
				cli.Logf("Incorrect start block index in the list: asked = %v, next in list = %v", firstIndex, blocks[i])
				return nil
			}
		}

		for ; i < len(blocks); i++ {
			blockIndex := blocks[i]

			cli.Logf("Validating block %v (%v/%v)...", blockIndex, i+1, len(blocks))
			srcState, err := srcStore.StateByIndex(blockIndex)
			if err != nil {
				cli.Logf("Failed to read old state for block %v: %v", blockIndex, err)
				failedBlocks = append(failedBlocks, blockIndex)
				continue
			}

			destState, err := destStore.StateByIndex(blockIndex)
			if err != nil {
				cli.Logf("Failed to read new state for block %v: %v", blockIndex, err)
				failedBlocks = append(failedBlocks, blockIndex)
				continue
			}

			oldState := utils.NewRecordingKVStoreReadOnly(srcState)
			newState := utils.NewRecordingKVStoreReadOnly(destState)

			succeded := true

			func() {
				defer func() {
					if err := recover(); err != nil {
						succeded = false
						failedBlocks = append(failedBlocks, blockIndex)
					}
				}()

				validateStatesEqual(oldState, newState, oldChainID, newChainID, firstIndex, blockIndex, short)
			}()

			cli.Logf("Block %v: %v", blockIndex, lo.Ternary(succeded, "OK", "FAILED"))
		}

		if len(failedBlocks) > 0 {
			cli.Logf("Validation failed for blocks: %v", failedBlocks)
		}
	default:
		cli.Logf("Invalid command line arguments: --blocks-list and --find-fail-block are mutually exclusive")
	}

	return nil
}

func validateStatesEqual(oldState old_kv.KVStoreReader, newState kv.KVStoreReader, oldChainID old_isc.ChainID, newChainID isc.ChainID, firstIndex, lastIndex uint32, short bool) {
	cli.DebugLogf("Validating states equality...\n")
	var oldStateContentStr, newStateContentStr string
	validation.GoAllAndWait(func() {
		oldStateContentStr = validation.OldStateContentToStr(oldState, oldChainID, firstIndex, lastIndex, short)
		cli.DebugLogf("Replacing old chain ID with constant placeholder for comparison...")
		oldStateContentStr = strings.Replace(oldStateContentStr, oldChainID.String(), "<chain-id>", -1)
		// TODO: review this
		cli.DebugLogf("Ignoring cross-chain agent ID for comparison...")
		const crossChainAgentID = "ContractAgentID(0x05204969, iota1pphx6hnmxqdqd2u4m59e7nvmcyulm3lfm58yex5gmud9qlt3v9crs9sah6m)"
		const crossChainAgentIDReplaced = "ContractAgentID(0x05204969, <chain-id>)"
		oldStateContentStr = strings.Replace(oldStateContentStr, crossChainAgentID, crossChainAgentIDReplaced, -1)
	}, func() {
		newStateContentStr = validation.NewStateContentToStr(newState, newChainID, firstIndex, lastIndex, short)
		cli.DebugLogf("Replacing new chain ID with constant placeholder for comparison...")
		newStateContentStr = strings.Replace(newStateContentStr, newChainID.String(), "<chain-id>", -1)
	})

	validation.EnsureEqual("states", oldStateContentStr, newStateContentStr)

	cli.ClearStatusBar()
	cli.DebugLogf("States are equal\n")
}

func findBlockWithIssue(firstBlockIndex, lastBlockIndex uint32, runValidation func(lastBlockIndex uint32) bool) {
	cli.Logf("Running validation for initial block range [%v, %v]", firstBlockIndex, lastBlockIndex)
	if runValidation(lastBlockIndex) {
		return
	}

	cli.Logf("Validation FAILED for block range [%v, %v]", firstBlockIndex, lastBlockIndex)

	cli.Logf("Trying to find block with issue...")
	lastFailedBlockIndex := lastBlockIndex
	searchRangeFirstBlockIndex := firstBlockIndex
	searchRangeLastBlockIndex := lastBlockIndex - 1

	for {
		cli.Logf("Searching in range [%v, %v]", searchRangeFirstBlockIndex, searchRangeLastBlockIndex)
		lastBlockIndex = (searchRangeLastBlockIndex + searchRangeFirstBlockIndex) / 2

		cli.Logf("Running validation for block range [%v, %v]", firstBlockIndex, lastBlockIndex)
		if runValidation(lastBlockIndex) {
			cli.Logf("Validation PASSED for block range [%v, %v]", searchRangeFirstBlockIndex, lastBlockIndex)
			if searchRangeLastBlockIndex == searchRangeFirstBlockIndex {
				break
			}

			searchRangeFirstBlockIndex = lastBlockIndex + 1
		} else {
			cli.Logf("Validation FAILED for block range [%v, %v]", searchRangeFirstBlockIndex, lastBlockIndex)
			lastFailedBlockIndex = lastBlockIndex
			if searchRangeLastBlockIndex == searchRangeFirstBlockIndex {
				break
			}

			searchRangeLastBlockIndex = lastBlockIndex - 1
		}

		if searchRangeLastBlockIndex < searchRangeFirstBlockIndex {
			break
		}
	}

	cli.Logf("Found block with issue: %v", lastFailedBlockIndex)
}

func readBlocksList(filePath string) ([]uint32, error) {
	text := lo.Must(os.ReadFile(filePath))
	lines := strings.Split(string(text), "\n")

	blockIndexes := make([]uint32, 0, len(lines))
	for lineIndex, line := range lines {
		commentPos := strings.Index(line, "#")
		if commentPos != -1 {
			line = line[:commentPos]
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		blockIndex, err := strconv.Atoi(line)
		if err != nil {
			return nil, fmt.Errorf("invalid block index in file %s: line %v: %s", filePath, lineIndex+1, line)
		}

		blockIndexes = append(blockIndexes, uint32(blockIndex))
	}

	slices.Sort(blockIndexes)
	blockIndexes = lo.Uniq(blockIndexes)

	return blockIndexes, nil
}
