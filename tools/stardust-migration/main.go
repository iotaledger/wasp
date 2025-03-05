// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"

	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/tools/stardust-migration/utils/cli"
	"github.com/samber/lo"
	cmd "github.com/urfave/cli/v2"
)

// NOTE: Every record type should be explicitly included in migration
// NOTE: All migration is node at once or just abandoned. There is no option to continue.
// TODO: Do we start from block 0 or N+1 where N last old block?
// TODO: Do we prune old block? Are we going to do migration from origin? If not, have we pruned blocks with old schemas?
// TODO: What to do with foundry prefixes?
// TODO: From where to get new chain ID?
// TODO: Need to migrate ALL trie roots to support tracing.
// TODO: New state draft might be huge, but it is stored in memory - might be an issue.

func main() {
	// For pprof profilings
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
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
								Aliases: []string{"i"},
								Usage:   "Specify block index to start from. If not specified, all blocks will be migrated starting from block 0.",
							},
							&cmd.BoolFlag{
								Name:  "hm-prefixes",
								Usage: "Replace original prefixes in new dabase with human-readable strings.",
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
								Name:  "hm-prefixes",
								Usage: "Replace original prefixes in new dabase with human-readable strings.",
							},
						},
						Before: processCommonFlags,
						Action: validateMigration,
					},
				},
			},
		},
	}

	lo.Must0(app.Run(os.Args))
}

func processCommonFlags(c *cmd.Context) error {
	if c.Bool("hm-prefixes") {
		cli.Logf("WARNING: Using human-readable prefixes\n")

		// NOTE: I've jsut did it for accounts for now
		accounts.PrefixAccountCoinBalances = "<coin_balances>"
		accounts.PrefixAccountWeiRemainder = "<wei_remainder>"
		accounts.L2TotalsAccount = "<l2_totals>"
		accounts.PrefixObjects = "<objects>"
		accounts.PrefixObjectsByCollection = "<objects_by_collection>"
		accounts.NoCollection = "<no_collection>"
		accounts.KeyNonce = "<nonce>"
		accounts.KeyCoinInfo = "<coin_info>"
		accounts.KeyObjectRecords = "<object_records>"
		accounts.KeyObjectOwner = "<object_owner>"
		accounts.KeyAllAccounts = "<all_accounts>"
	}

	return nil
}
