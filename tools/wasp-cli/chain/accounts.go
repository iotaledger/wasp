package chain

import (
	"context"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/clients/apiextensions"
	"github.com/iotaledger/wasp/clients/chainclient"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/gas"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/wallet"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/tools/wasp-cli/util"
	"github.com/iotaledger/wasp/tools/wasp-cli/waspcmd"
)

func initListAccountsCmd() *cobra.Command {
	var node string
	var chain string

	cmd := &cobra.Command{
		Use:   "list-accounts",
		Short: "List L2 accounts",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			node = waspcmd.DefaultWaspNodeFallback(node)
			chain = defaultChainFallback(chain)

			client := cliclients.WaspClient(node)
			chainID := config.GetChain(chain)

			accountList, _, err := client.CorecontractsApi.AccountsGetAccounts(context.Background(), chainID.String()).Execute() //nolint:bodyclose // false positive
			log.Check(err)

			log.Printf("Total %d account(s) in chain %s\n", len(accountList.Accounts), config.GetChain(chain).String())

			header := []string{"agentid"}
			rows := make([][]string, len(accountList.Accounts))
			for i, account := range accountList.Accounts {
				rows[i] = []string{account}
			}
			log.PrintTable(header, rows)
		},
	}

	waspcmd.WithWaspNodeFlag(cmd, &node)
	withChainFlag(cmd, &chain)
	return cmd
}

func initBalanceCmd() *cobra.Command {
	var node string
	var chain string
	cmd := &cobra.Command{
		Use:   "balance [<agentid>]",
		Short: "Show the L2 balance of the given L2 account (default: own account, `common`: chain common account)",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			node = waspcmd.DefaultWaspNodeFallback(node)
			chain = defaultChainFallback(chain)
			chainID := config.GetChain(chain)
			agentID := util.AgentIDFromArgs(args, chainID)
			client := cliclients.WaspClient(node)

			balance, _, err := client.CorecontractsApi.AccountsGetAccountBalance(context.Background(), chainID.String(), agentID.String()).Execute() //nolint:bodyclose // false positive
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

	waspcmd.WithWaspNodeFlag(cmd, &node)
	withChainFlag(cmd, &chain)
	return cmd
}

func initAccountNFTsCmd() *cobra.Command {
	var node string
	var chain string
	cmd := &cobra.Command{
		Use:   "nfts [<agentid>|common]",
		Short: "Show NFTs owned by a given account (default: own account, `common`: chain common account)",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			node = waspcmd.DefaultWaspNodeFallback(node)
			chain = defaultChainFallback(chain)
			chainID := config.GetChain(chain)
			agentID := util.AgentIDFromArgs(args, chainID)
			client := cliclients.WaspClient(node)

			nfts, _, err := client.CorecontractsApi.
				AccountsGetAccountNFTIDs(context.Background(), chainID.String(), agentID.String()).
				Execute() //nolint:bodyclose // false positive
			log.Check(err)

			for _, nftID := range nfts.NftIds {
				log.Printf("%s\n", nftID)
			}
		},
	}

	waspcmd.WithWaspNodeFlag(cmd, &node)
	withChainFlag(cmd, &chain)
	return cmd
}

// baseTokensForDepositFee calculates the amount of tokens needed to pay for a deposit
func baseTokensForDepositFee(client *apiclient.APIClient, chain string) uint64 {
	callGovView := func(viewName string) dict.Dict {
		result, _, err := client.ChainsApi.CallView(context.Background(), config.GetChain(chain).String()).
			ContractCallViewRequest(apiclient.ContractCallViewRequest{
				ContractName: governance.Contract.Name,
				FunctionName: viewName,
			}).Execute() //nolint:bodyclose // false positive
		log.Check(err)

		resultDict, err := apiextensions.APIJsonDictToDict(*result)
		log.Check(err)
		return resultDict
	}

	feePolicyBytes := callGovView(governance.ViewGetFeePolicy.Name).Get(governance.ParamFeePolicyBytes)
	feePolicy := gas.MustFeePolicyFromBytes(feePolicyBytes)

	gasLimitsBytes := callGovView(governance.ViewGetGasLimits.Name).Get(governance.ParamGasLimitsBytes)
	gasLimits, err := gas.LimitsFromBytes(gasLimitsBytes)
	log.Check(err)

	// assumes deposit fee == minGasPerRequest fee
	return feePolicy.FeeFromGas(gasLimits.MinGasPerRequest)
}

func initDepositCmd() *cobra.Command {
	var adjustStorageDeposit bool
	var node string
	var chain string

	cmd := &cobra.Command{
		Use:   "deposit [<agentid>] <token-id>:<amount>, [<token-id>:amount ...]",
		Short: "Deposit L1 funds into the given L2 account (default: own account, `common`: chain common account)",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			node = waspcmd.DefaultWaspNodeFallback(node)
			chain = defaultChainFallback(chain)

			chainID := config.GetChain(chain)
			if strings.Contains(args[0], ":") {
				// deposit to own agentID
				tokens := util.ParseFungibleTokens(util.ArgsToFungibleTokensStr(args))

				util.WithSCTransaction(config.GetChain(chain), node, func() (*iotago.Transaction, error) {
					client := cliclients.WaspClient(node)

					return cliclients.SCClient(client, chainID, accounts.Contract.Hname()).PostRequest(
						accounts.FuncDeposit.Name,
						chainclient.PostRequestParams{
							Transfer:                 tokens,
							AutoAdjustStorageDeposit: adjustStorageDeposit,
						},
					)
				})
			} else {
				// deposit to some other agentID
				agentID := util.AgentIDFromString(args[0], chainID)
				tokens := util.ParseFungibleTokens(util.ArgsToFungibleTokensStr(args[1:]))

				allowance := tokens.Clone()

				{
					// adjust allowance to leave enough for fee if needed
					client := cliclients.WaspClient(node)
					feeNeeded := baseTokensForDepositFee(client, chain)
					senderAgentID := isc.NewAgentID(wallet.Load().Address())
					senderOnChainBalance, _, err := client.CorecontractsApi.AccountsGetAccountBalance(context.Background(), chainID.String(), senderAgentID.String()).Execute() //nolint:bodyclose // false positive
					log.Check(err)
					senderOnChainBaseTokens, err := strconv.ParseUint(senderOnChainBalance.BaseTokens, 10, 64)
					log.Check(err)

					if senderOnChainBaseTokens < feeNeeded {
						allowance.Spend(isc.NewAssetsBaseTokens(feeNeeded - senderOnChainBaseTokens))
					}
				}

				util.WithSCTransaction(config.GetChain(chain), node, func() (*iotago.Transaction, error) {
					client := cliclients.WaspClient(node)

					return cliclients.SCClient(client, chainID, accounts.Contract.Hname()).PostRequest(
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
	waspcmd.WithWaspNodeFlag(cmd, &node)
	withChainFlag(cmd, &chain)

	return cmd
}
