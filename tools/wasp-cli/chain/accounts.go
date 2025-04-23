package chain

import (
	"context"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/clients/apiextensions"
	"github.com/iotaledger/wasp/clients/chainclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
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
			agentID := util.AgentIDFromArgs(args)
			ctx := context.Background()
			client := cliclients.WaspClientWithVersionCheck(ctx, node)

			balance, _, err := client.CorecontractsAPI.AccountsGetAccountBalance(ctx, agentID.String()).Execute() //nolint:bodyclose // false positive
			log.Check(err)

			header := []string{"token", "amount"}
			rows := make([][]string, len(balance.NativeTokens)+1)

			rows[0] = []string{"base", balance.BaseTokens}
			for k, v := range balance.NativeTokens {
				rows[k+1] = []string{v.CoinType, v.Balance}
			}

			log.PrintTable(header, rows)
		},
	}

	waspcmd.WithWaspNodeFlag(cmd, &node)
	withChainFlag(cmd, &chain)
	return cmd
}

func initAccountObjectsCmd() *cobra.Command {
	var node string
	var chain string
	cmd := &cobra.Command{
		Use:   "objects [<agentid>|common]",
		Short: "Show non-coin objects owned by a given account (default: own account, `common`: chain common account)",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			node = waspcmd.DefaultWaspNodeFallback(node)
			chain = defaultChainFallback(chain)
			agentID := util.AgentIDFromArgs(args)
			ctx := context.Background()
			client := cliclients.WaspClientWithVersionCheck(ctx, node)

			nfts, _, err := client.CorecontractsAPI.
				AccountsGetAccountNFTIDs(ctx, agentID.String()).
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
//
//nolint:unused
func baseTokensForDepositFee(client *apiclient.APIClient) coin.Value {
	callGovView := func(viewName string) isc.CallResults {
		apiResult, _, err := client.ChainsAPI.CallView(context.Background()).
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
	return feePolicy.FeeFromGas(gasLimits.MinGasPerRequest, nil, parameters.BaseTokenDecimals)
}

func initDepositCmd() *cobra.Command {
	var printReceipt bool
	var node string
	var chain string

	cmd := &cobra.Command{
		Use:   "deposit [<agentid>] <token-id1>|<amount1>, [<token-id2>|<amount2> ...]",
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
			var res *iotajsonrpc.IotaTransactionBlockResponse
			var err error
			if strings.Contains(args[0], "|") {
				// deposit to own agentID
				tokens := util.ParseFungibleTokens(util.ArgsToFungibleTokensStr(args))
				allowance := tokens.Clone()
				allowance.SetBaseTokens(allowance.BaseTokens())

				res = util.WithSCTransaction(ctx, client, func() (*iotajsonrpc.IotaTransactionBlockResponse, error) {
					return cliclients.ChainClient(client, chainID).PostRequest(ctx,
						accounts.FuncDeposit.Message(),
						chainclient.PostRequestParams{
							Transfer:    tokens,
							Allowance:   allowance,
							GasBudget:   iotaclient.DefaultGasBudget,
							L2GasBudget: isc.Million,
						},
					)
				})

				log.Printf("Posted TX: %s\n", res.Digest)
			} else {
				// deposit to some other agentID
				agentID := util.AgentIDFromString(args[0])
				tokens := util.ParseFungibleTokens(util.ArgsToFungibleTokensStr(args[1:]))
				allowance := tokens.Clone()
				allowance.SetBaseTokens(allowance.BaseTokens())

				res = util.WithSCTransaction(ctx, client, func() (*iotajsonrpc.IotaTransactionBlockResponse, error) {
					return cliclients.ChainClient(client, chainID).PostRequest(
						ctx,
						accounts.FuncTransferAllowanceTo.Message(agentID),
						chainclient.PostRequestParams{
							Transfer:    tokens,
							Allowance:   allowance,
							GasBudget:   iotaclient.DefaultGasBudget,
							L2GasBudget: isc.Million,
						},
					)
				})

				log.Check(err)
			}
			log.Printf("Posted TX: %s\n", res.Digest)
			if printReceipt {
				log.Printf("L1 Gas Fee: %d\n", res.Effects.Data.GasFee())
				ref, err := res.GetCreatedObjectInfo("request", "Request")
				log.Check(err)
				log.Printf("Requet ID: %s\n", ref.ObjectID.String())
				receipt, _, err := client.ChainsAPI.
					GetReceipt(ctx, ref.ObjectID.String()).
					Execute() //nolint:bodyclose // false positive

				log.Check(err)
				util.LogReceipt(*receipt, 0)
			}
		},
	}

	cmd.Flags().BoolVarP(&printReceipt, "print-receipt", "p", false, "print tx recetip")
	waspcmd.WithWaspNodeFlag(cmd, &node)
	withChainFlag(cmd, &chain)

	return cmd
}
