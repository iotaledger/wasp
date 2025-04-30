// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"math"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	old_kv "github.com/nnikolash/wasp-types-exported/packages/kv"
	old_collections "github.com/nnikolash/wasp-types-exported/packages/kv/collections"
	old_state "github.com/nnikolash/wasp-types-exported/packages/state"
	old_indexedstore "github.com/nnikolash/wasp-types-exported/packages/state/indexedstore"
	old_blocklog "github.com/nnikolash/wasp-types-exported/packages/vm/core/blocklog"
	"github.com/samber/lo"
	"github.com/slack-go/slack"
	cmd "github.com/urfave/cli/v2"

	bcs "github.com/iotaledger/bcs-go"
	old_iotago "github.com/iotaledger/iota.go/v3"

	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/trie"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/tools/stardust-migration/bot"
	"github.com/iotaledger/wasp/tools/stardust-migration/stateaccess/newstate"
	"github.com/iotaledger/wasp/tools/stardust-migration/stateaccess/oldstate"
	"github.com/iotaledger/wasp/tools/stardust-migration/utils/cli"
	"github.com/iotaledger/wasp/tools/stardust-migration/utils/db"
)

func main() {
	// For pprof profilings
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()

	bot.Get().PostWelcomeMessage(fmt.Sprintf("*A new migration has been started.* %s", time.Now().String()))

	// For Slack notifications

	defer func() { //catch or finally
		if err := recover(); err != nil { //catch
			errorStr := fmt.Sprintf(":collision: *Migration panicked!*\nError: %v\nStack: %v\n <!here> ", err, string(debug.Stack()))
			bot.Get().PostMessage(errorStr, slack.MsgOptionLinkNames(true))
			log.Println(errorStr)
			os.Exit(1)
		}
	}()

	app := &cmd.App{
		Name:        "stardust-migrator",
		Description: "Stardust migration tool",
		Commands: []*cmd.Command{
			{
				Name: "migrate",
				Subcommands: []*cmd.Command{
					{
						Name:      "all-states",
						ArgsUsage: "<src-chain-db-dir> <dest-chain-db-dir>",
						Flags: []cmd.Flag{
							&cmd.Uint64Flag{
								Name:    "from-index",
								Aliases: []string{"i", "f", "from-block", "from"},
								Usage:   "Specify block index to start from. If not specified, all blocks will be migrated starting from block 0.",
							},
							&cmd.Uint64Flag{
								Name:    "to-index",
								Aliases: []string{"t", "to-block", "to"},
								Usage:   "Specify block index to migrate last. If not specified, all blocks will be migrated until last.",
							},
							&cmd.BoolFlag{
								Name:    "skip-load",
								Aliases: []string{"no-load"},
								Usage:   "Do not pre-load full state at the start of migration, when using option '--from-index' / '-i'. WARNING: This will result in a BROKEN migrated db.",
							},
							&cmd.BoolFlag{
								Name:    "continue",
								Aliases: []string{"c"},
								Usage:   "Continue migration from the last block in the destination database.",
							},
							&cmd.BoolFlag{
								Name:    "dummy-chain-owner",
								Aliases: []string{"o"},
								Usage:   "Disables reading preparation config from file and instead uses dummy config.",
							},
							&cmd.BoolFlag{
								Name:  "no-state-cache",
								Usage: "Disable reading pre-saved in-memory states from files. This forces loading entire latest state from DB (may take a lot of time).",
							},
							&cmd.BoolFlag{
								Name:  "periodic-state-save",
								Usage: "Save state every 20000 blocks. This will slow down the migration, but might allow to continue it later in case of unexpected interruption.",
							},
							&cmd.BoolFlag{
								Name:  "refcount-cache",
								Usage: "Enable storing refcounts in memory. Otherwise they will always be directly written to DB. This will speed up the migration, but will use huge amount of memory.",
							},
							&cmd.BoolFlag{
								Name:    "dry-run",
								Aliases: []string{"d"},
								Usage:   "Do not write destination database.",
							},
							&cmd.BoolFlag{
								Name:  "hm-prefixes",
								Usage: "Replace original prefixes in new dabase with human-readable strings.",
							},
							&cmd.StringFlag{
								Name:  "debug-dest-key",
								Usage: "Print stack when destination db key CONTAIN given hex string (works as AND with --debug-dest-value).",
							},
							&cmd.StringFlag{
								Name:  "chain-owner",
								Usage: "Sets the chain owner address of the to-be migrated chain",
							},
							&cmd.StringFlag{
								Name:  "debug-dest-value",
								Usage: "Print stack when destination db value CONTAIN given hex string (works as AND with --debug-dest-key).",
							},
							&cmd.StringFlag{
								Name:  "debug-filter-trace",
								Usage: "Print stacktrace only if it contains given string (used with --debug-dest-key and --debug-dest-value).",
							},
							&cmd.BoolFlag{
								Name:    "print-block-idx",
								Aliases: []string{"print-block-index", "print-idx"},
								Usage:   "Print block index for each block.",
							},
						},
						Before: processCommonFlags,
						Action: migrateAllStates,
					},
				},
			},
			{
				Name: "validate",
				Subcommands: []*cmd.Command{
					{
						Name:      "migration",
						ArgsUsage: "<src-chain-db-dir> <dest-chain-db-dir>",
						Flags: []cmd.Flag{
							&cmd.BoolFlag{
								Name:    "short",
								Aliases: []string{"s"},
								Usage:   "Skip some of long validations.",
							},
							&cmd.BoolFlag{
								Name:    "no-parallel",
								Aliases: []string{"p"},
								Usage:   "Do not run validation in parallel.",
							},
							&cmd.BoolFlag{
								Name:  "hm-prefixes",
								Usage: "Replace original prefixes in new dabase with human-readable strings.",
							},
							&cmd.Uint64Flag{
								Name:    "from-index",
								Aliases: []string{"i", "from-block", "from"},
								Usage:   "Specify block index to start from. This is used as hint in blocklog migration for cases, when database was generated not from first block.",
							},
							&cmd.Uint64Flag{
								Name:    "to-index",
								Aliases: []string{"t", "to-block", "to"},
								Value:   math.MaxUint64,
								Usage:   "Specify block to validate. If not specified, latest available block is validated.",
							},
							&cmd.StringFlag{
								Name:    "blocks-list",
								Aliases: []string{"l"},
								Usage:   "Specify file with list of blocks to validate. It's like running the tool multiple times with -t/--to-index option. See blocks_to_validate.txt file as example of syntax.",
							},
							&cmd.BoolFlag{
								Name:  "no-hashing",
								Usage: "Do not hash data before comparing. Will produce bigger but more user-friendly output.",
							},
							&cmd.BoolFlag{
								Name:    "find-fail-block",
								Aliases: []string{"f"},
								Usage:   "Find the first block where validation fails using binary search.",
							},
						},
						Before: processCommonFlags,
						Action: validateMigration,
					},
				},
			},
			{
				Name:      "create-index",
				ArgsUsage: "path/to/rebased/db path/for/index/db",
				Action:    createIndex,
				Flags: []cmd.Flag{
					&cmd.BoolFlag{
						Name:    "parallel",
						Aliases: []string{"p"},
						Usage:   "Create index in parallel.",
					},
				},
			},
			{
				Name:      "webapi-validate",
				ArgsUsage: "http://stardust-isc:9090 http://rebased-isc:9090",
				Flags: []cmd.Flag{
					&cmd.Uint64Flag{
						Name:    "from-index",
						Aliases: []string{"i", "f", "from-block", "from"},
						Usage:   "Specify block index to start from. This is used as hint in blocklog migration for cases, when database was generated not from first block.",
					},
					&cmd.Uint64Flag{
						Name:    "to-index",
						Aliases: []string{"t", "to-block", "to"},
						Usage:   "Specify block to validate. If not specified, latest available block is validated.",
					},
				},
				Action: validateWebAPI,
			},
			{
				Name:      "account-dump-validate",
				ArgsUsage: "stardust-isc-account-dump.json rebased-isc-account-dump.json",
				Flags:     []cmd.Flag{},
				Action:    validateAccountDumps,
			},
			{
				Name: "search",
				Subcommands: []*cmd.Command{
					searchCmd("iscmagic-allowance", searchISCMagicAllowance),
					searchCmd("nft", searchNFT, IncludeDeletions()),
					searchCmd("block-keep-amount-change", searchBlockKeepAmountNot10000),
					searchCmd("foundry", searchFoundies),
					searchCmd("native-token", searchNativeTokens),
					searchCmd("strange-native-token", searchStrangeNativeTokenRecords, IncludeDeletions()),
					searchCmd("key", searchKey, ArgsUsage("<key-hex>"), IncludeDeletions()),
					searchCmd("gas-fee-policy", searchGasFeePolicyChange),
					searchCmd("gas-budget-exceeded", searchGasBudgetExceeded),
					searchCmd("nicole-coin", searchNicoleCoin),
					searchCmd("cross-chain", searchCrossChain),
				},
			},
			{
				Name:      "get-block-muts",
				ArgsUsage: "<chain-db-dir> <block-index>",
				Action:    getBlockMuts,
			},
			{
				Name:      "get-state-value",
				ArgsUsage: "<chain-db-dir> <state-index> <key-hex>",
				Action:    getStateValue,
			},
		},
	}

	programCtx, _ := signal.NotifyContext(context.Background(), os.Interrupt)
	lo.Must0(app.RunContext(programCtx, os.Args))
}

func searchCmd(entityName string, f SearchFuncConstructor, opts ...SearchOption) *cmd.Command {
	options := SearchOptions{}
	for _, opt := range opts {
		opt(&options)
	}

	if options.ArgsUsage == "" {
		options.ArgsUsage = "[<custom_arg1>] [<custom_arg2>] ..."
	}

	return &cmd.Command{
		Name:      entityName,
		ArgsUsage: "<chain-db-dir> " + options.ArgsUsage,
		Flags: []cmd.Flag{
			&cmd.Uint64Flag{
				Name:    "from-index",
				Aliases: []string{"i", "f", "from-block", "from"},
				Usage:   "Start search from this block index.",
			},
			&cmd.Uint64Flag{
				Name:    "to-index",
				Aliases: []string{"t", "to-block", "to"},
				Usage:   "Stop search at this block index.",
			},
			&cmd.BoolFlag{
				Name:    "all",
				Aliases: []string{"a"},
				Usage:   "Find all occurrences (by default, only first is printed and search stops).",
			},
			&cmd.UintFlag{
				Name:        "parallel",
				Aliases:     []string{"p"},
				Usage:       "Number of parallel threads to use for search.",
				Value:       uint(runtime.NumCPU() * 2),
				DefaultText: fmt.Sprintf("%v", runtime.NumCPU()*2),
			},
		},
		Before: processCommonFlags,
		Action: search(entityName, f, options),
	}
}

func getBlockMuts(c *cmd.Context) error {
	chainDBDir := c.Args().Get(0)
	blockIndexStr := c.Args().Get(1)

	blockIndex := lo.Must(strconv.Atoi(blockIndexStr))

	kvs := db.ConnectOld(chainDBDir)
	store := old_indexedstore.New(old_state.NewStoreWithUniqueWriteMutex(kvs))

	b := lo.Must(store.BlockByIndex(uint32(blockIndex)))
	for i, m := range b.Mutations().Sets {
		fmt.Printf("SET %x: %x\n", i, m)
	}
	for i, m := range b.Mutations().Dels {
		fmt.Printf("DEL %x: %x\n", i, m)
	}

	return nil
}

func getStateValue(c *cmd.Context) error {
	chainDBDir := c.Args().Get(0)
	stateIndexStr := c.Args().Get(1)
	keyHex := c.Args().Get(2)

	stateIndex := lo.Must(strconv.Atoi(stateIndexStr))
	if !strings.HasPrefix(keyHex, "0x") {
		keyHex = "0x" + keyHex
	}
	key := lo.Must(cryptolib.DecodeHex(keyHex))

	kvs := db.ConnectOld(chainDBDir)
	store := old_indexedstore.New(old_state.NewStoreWithUniqueWriteMutex(kvs))

	s, err := store.StateByIndex(uint32(stateIndex))
	if err != nil {
		return err
	}

	fmt.Printf("Key: %x\n", key)
	fmt.Printf("Value: %x\n", s.Get(old_kv.Key(key)))
	fmt.Printf("Has value: %v\n", s.Has(old_kv.Key(key)))

	return nil
}

func processCommonFlags(c *cmd.Context) error {
	if c.Bool("hm-prefixes") {
		cli.Logf("WARNING: Using human-readable prefixes\n")

		trie.KeyMaxLength = 512

		accounts.PrefixAccountCoinBalances = "<coin_balances>"
		accounts.PrefixAccountWeiRemainder = "<wei_remainder>"
		accounts.L2TotalsAccount = "<l2_totals>"
		accounts.PrefixObjects = "<objects>"
		//accounts.PrefixObjectsByCollection = "<objects_by_collection>"
		//accounts.NoCollection = "<no_collection>"
		accounts.KeyNonce = "<nonce>"
		accounts.KeyCoinInfo = "<coin_info>"
		//accounts.KeyObjectRecords = "<object_records>"
		accounts.KeyObjectOwner = "<object_owner>"
		accounts.KeyAllAccounts = "<all_accounts>"

		blocklog.PrefixBlockRegistry = "<block_registry>"
		blocklog.PrefixRequestEvents = "<request_events>"
		blocklog.PrefixRequestLookupIndex = "<request_lookup_index>"
		blocklog.PrefixRequestReceipts = "<request_receipts>"
	}

	return nil
}

func GetAnchorOutput(chainState old_kv.KVStoreReader) *old_iotago.AliasOutput {
	contractState := oldstate.GetContactStateReader(chainState, old_blocklog.Contract.Hname())

	registry := old_collections.NewArrayReadOnly(contractState, old_blocklog.PrefixBlockRegistry)
	if registry.Len() == 0 {
		panic("Block registry is empty")
	}

	blockInfoBytes := registry.GetAt(registry.Len() - 1)

	var blockInfo old_blocklog.BlockInfo
	lo.Must0(blockInfo.Read(bytes.NewReader(blockInfoBytes)))

	return blockInfo.PreviousAliasOutput.GetAliasOutput()
}

func GetStateAnchor(chainState kv.KVStoreReader) *isc.StateAnchor {
	contractState := newstate.GetContactStateReader(chainState, blocklog.Contract.Hname())

	registry := collections.NewArrayReadOnly(contractState, blocklog.PrefixBlockRegistry)
	if registry.Len() == 0 {
		panic("Block registry is empty")
	}

	blockInfoBytes := registry.GetAt(registry.Len() - 1)
	blockInfo := bcs.MustUnmarshal[blocklog.BlockInfo](blockInfoBytes)

	return blockInfo.PreviousAnchor
}

func GetAnchorObject(chainState kv.KVStoreReader) *iscmove.Anchor {
	return GetStateAnchor(chainState).Anchor().Object
}
