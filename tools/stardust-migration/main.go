// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	"time"

	old_kv "github.com/nnikolash/wasp-types-exported/packages/kv"
	old_collections "github.com/nnikolash/wasp-types-exported/packages/kv/collections"
	old_blocklog "github.com/nnikolash/wasp-types-exported/packages/vm/core/blocklog"
	"github.com/samber/lo"
	"github.com/slack-go/slack"
	cmd "github.com/urfave/cli/v2"

	bcs "github.com/iotaledger/bcs-go"
	old_iotago "github.com/iotaledger/iota.go/v3"

	"github.com/iotaledger/wasp/clients/iscmove"
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
		Name: "Stardust migration tool",
		Commands: []*cmd.Command{
			{
				Name: "migrate",
				Subcommands: []*cmd.Command{
					{
						Name:      "single-state",
						ArgsUsage: "<src-chain-db-dir> <dest-chain-db-dir>",
						Flags: []cmd.Flag{
							&cmd.Uint64Flag{
								Name:    "index",
								Aliases: []string{"i"},
								Usage:   "Specify block index to migrate. If not specified, latest state will be migrated.",
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
						},
						Before: processCommonFlags,
						Action: migrateSingleState,
					},
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
								Aliases: []string{"i", "f", "from-block", "from"},
								Usage:   "Specify block index to start from. This is used as hint in blocklog migration for cases, when database was generated not from first block.",
							},
							&cmd.Uint64Flag{
								Name:    "to-index",
								Aliases: []string{"t", "to-block", "to"},
								Usage:   "Specify block to validate. If not specified, latest available block is validated.",
							},
							&cmd.BoolFlag{
								Name:  "no-hashing",
								Usage: "Do not hash data before comparing. Will produce bigger but more user-friendly output.",
							},
						},
						Before: processCommonFlags,
						Action: validateMigration,
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
				Name: "search",
				Subcommands: []*cmd.Command{
					searchCmd("iscmagic-allowance", searchISCMagicAllowance),
					searchCmd("nft", searchNFT),
				},
			},
		},
	}

	programCtx, _ := signal.NotifyContext(context.Background(), os.Interrupt)
	lo.Must0(app.RunContext(programCtx, os.Args))
}

func searchCmd(entityName string, f StateContainsTargetCheckFunc) *cmd.Command {
	return &cmd.Command{
		Name:      entityName,
		ArgsUsage: "<chain-db-dir>",
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
		Action: search(entityName, f),
	}
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
