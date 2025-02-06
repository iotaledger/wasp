package chain

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/clients/apiextensions"
	"github.com/iotaledger/wasp/clients/chainclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/governance"

	"github.com/iotaledger/wasp/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/tools/wasp-cli/util"
	"github.com/iotaledger/wasp/tools/wasp-cli/waspcmd"
)

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
			ctx := context.Background()
			client := cliclients.WaspClientWithVersionCheck(ctx, node)

			balance, _, err := client.CorecontractsAPI.AccountsGetAccountBalance(ctx, chainID.String(), agentID.String()).Execute() //nolint:bodyclose // false positive
			log.Check(err)

			header := []string{"token", "amount"}
			rows := make([][]string, len(balance.NativeTokens)+1)

			rows[0] = []string{"base", balance.BaseTokens}
			for k, v := range balance.NativeTokens {
				rows[k+1] = []string{v.CoinType.GetS(), v.Balance}
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
			ctx := context.Background()
			client := cliclients.WaspClientWithVersionCheck(ctx, node)

			nfts, _, err := client.CorecontractsAPI.
				AccountsGetAccountNFTIDs(ctx, chainID.String(), agentID.String()).
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
func baseTokensForDepositFee(client *apiclient.APIClient, chain string) coin.Value {
	callGovView := func(viewName string) isc.CallResults {
		apiResult, _, err := client.ChainsAPI.CallView(context.Background(), config.GetChain(chain).String()).
			ContractCallViewRequest(apiclient.ContractCallViewRequest{
				ContractName: governance.Contract.Name,
				FunctionName: viewName,
			}).Execute() //nolint:bodyclose // false positive
		log.Check(err)

		result, err := apiextensions.APIResultToCallArgs(apiResult)
		log.Check(err)
		return result
	}

	r := callGovView(governance.ViewGetFeePolicy.Name)
	feePolicy, err := governance.ViewGetFeePolicy.DecodeOutput(r)
	log.Check(err)

	if feePolicy.GasPerToken.HasZeroComponent() {
		return 0
	}

	r = callGovView(governance.ViewGetGasLimits.Name)
	gasLimits, err := governance.ViewGetGasLimits.DecodeOutput(r)
	log.Check(err)

	// assumes deposit fee == minGasPerRequest fee
	return feePolicy.FeeFromGas(gasLimits.MinGasPerRequest, nil, parameters.Decimals)
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
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*1000)
			defer cancel()

			client := cliclients.WaspClientWithVersionCheck(ctx, node)

			util.TryMergeAllCoins(ctx)

			if strings.Contains(args[0], "|") {
				// deposit to own agentID
				tokens := util.ParseFungibleTokens(util.ArgsToFungibleTokensStr(args))
				allowance := isc.NewAssets(tokens.BaseTokens() / 10)

				util.WithSCTransaction(ctx, client, config.GetChain(chain), func() (*iotajsonrpc.IotaTransactionBlockResponse, error) {
					return cliclients.ChainClient(client, chainID).PostRequest(ctx,
						accounts.FuncDeposit.Message(),
						chainclient.PostRequestParams{
							Transfer:  tokens,
							Allowance: allowance,
						},
					)
				})
			} else {
				// deposit to some other agentID
				agentID := util.AgentIDFromString(args[0], chainID)
				tokens := util.ParseFungibleTokens(util.ArgsToFungibleTokensStr(args[1:]))

				/*
					allowance := tokens.Clone()
					{
						// adjust allowance to leave enough for fee if needed
						feeNeeded := coin.Value(10000)

						senderAgentID := isc.NewAddressAgentID(wallet.Load().Address())
						senderOnChainBalance, _, err := client.CorecontractsAPI.AccountsGetAccountBalance(context.Background(), chainID.String(), senderAgentID.String()).Execute() //nolint:bodyclose // false positive
						log.Check(err)
						senderOnChainBaseTokens, err := strconv.ParseUint(senderOnChainBalance.BaseTokens, 10, 64)
						log.Check(err)

						if coin.Value(senderOnChainBaseTokens) < feeNeeded {
							allowance.Spend(isc.NewAssets(feeNeeded - coin.Value(senderOnChainBaseTokens)))
						}
					}*/

				allowance := isc.NewAssets(tokens.BaseTokens() - 10000)

				res, err := cliclients.ChainClient(client, chainID).PostRequest(
					ctx,
					accounts.FuncTransferAllowanceTo.Message(agentID),
					chainclient.PostRequestParams{
						Transfer:  tokens,
						Allowance: allowance,
					},
				)

				log.Check(err)
				fmt.Printf("Posted TX: %s\n", res.Digest)

			}

		},
	}

	cmd.Flags().BoolVarP(&adjustStorageDeposit, "adjust-storage-deposit", "s", false, "adjusts the amount of base tokens sent, if it's lower than the min storage deposit required")
	waspcmd.WithWaspNodeFlag(cmd, &node)
	withChainFlag(cmd, &chain)

	return cmd
}
