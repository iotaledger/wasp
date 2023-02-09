package chain

import (
	"context"
	"strings"

	"github.com/spf13/cobra"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/clients/chainclient"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/wallet"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/tools/wasp-cli/util"
	"github.com/iotaledger/wasp/tools/wasp-cli/waspcmd"
)

func initListAccountsCmd() *cobra.Command {
	var node string
	cmd := &cobra.Command{
		Use:   "list-accounts",
		Short: "List L2 accounts",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			node = waspcmd.DefaultSingleNodeFallback(node)
			client := cliclients.WaspClient(node)
			chainID := config.GetCurrentChainID()

			accountList, _, err := client.CorecontractsApi.AccountsGetAccounts(context.Background(), chainID.String()).Execute()
			log.Check(err)

			log.Printf("Total %d account(s) in chain %s\n", len(accountList.Accounts), config.GetCurrentChainID().String())

			header := []string{"agentid"}
			rows := make([][]string, len(accountList.Accounts))
			for i, account := range accountList.Accounts {
				rows[i] = []string{account}
				i++
			}
			log.PrintTable(header, rows)
		},
	}

	waspcmd.WithSingleWaspNodesFlag(cmd, &node)
	return cmd
}

func initBalanceCmd() *cobra.Command {
	var node string
	cmd := &cobra.Command{
		Use:   "balance [<agentid>]",
		Short: "Show the L2 balance of the given account",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			node = waspcmd.DefaultSingleNodeFallback(node)
			var agentID isc.AgentID
			if len(args) == 0 {
				agentID = isc.NewAgentID(wallet.Load().Address())
			} else {
				var err error
				agentID, err = isc.NewAgentIDFromString(args[0])
				log.Check(err)
			}

			client := cliclients.WaspClient(node)
			chainID := config.GetCurrentChainID()
			balance, _, err := client.CorecontractsApi.AccountsGetAccountBalance(context.Background(), chainID.String(), agentID.String()).Execute()

			log.Check(err)

			header := []string{"token", "amount"}
			rows := make([][]string, len(balance.NativeTokens)+1)

			rows[0] = []string{"base", balance.BaseTokens}
			for k, v := range balance.NativeTokens {
				rows[k+1] = []string{v.Id, v.Amount}
			}

			log.PrintTable(header, rows)
		},
	}

	waspcmd.WithSingleWaspNodesFlag(cmd, &node)
	return cmd
}

func initDepositCmd() *cobra.Command {
	var adjustStorageDeposit bool
	var node string

	cmd := &cobra.Command{
		Use:   "deposit [<agentid>] <token-id>:<amount>, [<token-id>:amount ...]",
		Short: "Deposit L1 funds into the given (default: your) L2 account",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			node = waspcmd.DefaultSingleNodeFallback(node)
			if strings.Contains(args[0], ":") {
				// deposit to own agentID
				tokens := util.ParseFungibleTokens(args)
				util.WithSCTransaction(config.GetCurrentChainID(), node, func() (*iotago.Transaction, error) {
					client := cliclients.WaspClient(node)

					return cliclients.SCClient(client, accounts.Contract.Hname()).PostRequest(
						accounts.FuncDeposit.Name,
						chainclient.PostRequestParams{
							Transfer:                 tokens,
							AutoAdjustStorageDeposit: adjustStorageDeposit,
						},
					)
				})
			} else {
				// deposit to some other agentID
				agentID, err := isc.NewAgentIDFromString(args[0])
				log.Check(err)
				tokensStr := strings.Split(strings.Join(args[1:], ""), ",")
				tokens := util.ParseFungibleTokens(tokensStr)
				allowance := tokens.Clone()

				util.WithSCTransaction(config.GetCurrentChainID(), node, func() (*iotago.Transaction, error) {
					client := cliclients.WaspClient(node)

					return cliclients.SCClient(client, accounts.Contract.Hname()).PostRequest(
						accounts.FuncTransferAllowanceTo.Name,
						chainclient.PostRequestParams{
							Args: dict.Dict{
								accounts.ParamAgentID: agentID.Bytes(),
							},
							Transfer:                 tokens,
							Allowance:                allowance,
							AutoAdjustStorageDeposit: adjustStorageDeposit,
						},
					)
				})
			}
		},
	}

	cmd.Flags().BoolVarP(&adjustStorageDeposit, "adjust-storage-deposit", "s", false, "adjusts the amount of base tokens sent, if it's lower than the min storage deposit required")
	waspcmd.WithSingleWaspNodesFlag(cmd, &node)

	return cmd
}
