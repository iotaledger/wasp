package chain

import (
	"context"
	"strings"
	"time"

	"github.com/samber/lo"
	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/v2/clients/apiclient"
	"github.com/iotaledger/wasp/v2/clients/apiextensions"
	"github.com/iotaledger/wasp/v2/clients/chainclient"
	"github.com/iotaledger/wasp/v2/clients/iotagraphql"
	"github.com/iotaledger/wasp/v2/packages/coin"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/parameters"
	"github.com/iotaledger/wasp/v2/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/v2/packages/vm/core/governance"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/cli/config"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/format"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/util"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/waspcmd"
)

func initBalanceCmd() *cobra.Command {
	var node string
	var chain string
	cmd := &cobra.Command{
		Use:   "balance [<agentid>]",
		Short: "Show the L2 balance of the given L2 account (default: own account, `common`: chain common account)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			node, err = waspcmd.DefaultWaspNodeFallback(node)
			if err != nil {
				return err
			}
			agentID, err := util.AgentIDFromArgs(args)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client := cliclients.WaspClientWithVersionCheck(ctx, node)

			balance, _, err := client.CorecontractsAPI.AccountsGetAccountBalance(ctx, agentID.String()).Execute() //nolint:bodyclose // false positive
			if err != nil {
				return err
			}

			coins := make([]format.ChainBalanceCoin, 0, len(balance.Coins)+1)
			coins = append(coins, format.ChainBalanceCoin{
				Token:  "base",
				Amount: balance.BaseTokens,
			})
			for _, v := range balance.Coins {
				if lo.Must(coin.IsBaseToken(v.CoinType)) {
					continue
				}
				coins = append(coins, format.ChainBalanceCoin{
					Token:  v.CoinType,
					Amount: v.Balance,
				})
			}

			output := format.ChainBalanceOutput{
				Coins: coins,
			}
			return format.FormatSuccess("chain_balance", output.ToMap())
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
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			node, err = waspcmd.DefaultWaspNodeFallback(node)
			if err != nil {
				return err
			}
			chain, err = defaultChainFallback(chain)
			if err != nil {
				return err
			}
			agentID, err := util.AgentIDFromArgs(args)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client := cliclients.WaspClientWithVersionCheck(ctx, node)

			assets, _, err := client.CorecontractsAPI.
				AccountsGetAccountBalance(ctx, agentID.String()).
				Execute() //nolint:bodyclose // false positive
			if err != nil {
				return err
			}

			for _, obj := range assets.Objects {
				log.Printf("%s: %s\n", obj.Type, obj.Id)
			}
			return nil
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
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			node, err = waspcmd.DefaultWaspNodeFallback(node)
			if err != nil {
				return err
			}
			chain, err = defaultChainFallback(chain)
			if err != nil {
				return err
			}
			chainID := config.GetChain(chain)

			ctx, cancel := context.WithTimeout(context.Background(), time.Second*1000)
			defer cancel()

			client := cliclients.WaspClientWithVersionCheck(ctx, node)

			util.TryManageCoinsAmount(ctx)
			var res *iotagraphql.ExecuteTransactionBlockResponse
			if strings.Contains(args[0], "|") {
				// deposit to own agentID
				var tokens *isc.Assets
				tokens, err = util.ParseFungibleTokens(util.ArgsToFungibleTokensStr(args))
				if err != nil {
					return err
				}
				allowance := tokens.Clone()
				allowance.SetBaseTokens(allowance.BaseTokens())

				res = util.WithSCTransaction(ctx, client, func() (*iotagraphql.ExecuteTransactionBlockResponse, error) {
					return cliclients.ChainClient(client, chainID).PostRequest(ctx,
						accounts.FuncDeposit.Message(),
						chainclient.PostRequestParams{
							Transfer:    tokens,
							Allowance:   allowance,
							GasBudget:   iotagraphql.DefaultGasBudget,
							L2GasBudget: isc.Million,
						},
					)
				})
			} else {
				// deposit to some other agentID
				var agentID isc.AgentID
				agentID, err = util.AgentIDFromString(args[0])
				if err != nil {
					return err
				}
				tokens, err := util.ParseFungibleTokens(util.ArgsToFungibleTokensStr(args[1:]))
				if err != nil {
					return err
				}
				allowance := tokens.Clone()
				allowance.SetBaseTokens(allowance.BaseTokens())

				res = util.WithSCTransaction(ctx, client, func() (*iotagraphql.ExecuteTransactionBlockResponse, error) {
					return cliclients.ChainClient(client, chainID).PostRequest(
						ctx,
						accounts.FuncTransferAllowanceTo.Message(agentID),
						chainclient.PostRequestParams{
							Transfer:    tokens,
							Allowance:   allowance,
							GasBudget:   iotagraphql.DefaultGasBudget,
							L2GasBudget: isc.Million,
						},
					)
				})
			}

			if printReceipt {
				if err := format.FormatSuccess("l1_gas_fee", map[string]interface{}{
					"amount": res.ExecuteTransactionBlock.Effects.GasFee(),
				}); err != nil {
					return err
				}
				ref, err := res.ExecuteTransactionBlock.Effects.GetCreatedObjectByName("request", "Request")
				if err != nil {
					return err
				}
				receipt, _, err := client.ChainsAPI.
					GetReceipt(ctx, ref.ObjectID.String()).
					Execute() //nolint:bodyclose // false positive
				if err != nil {
					return err
				}
				util.LogReceipt(*receipt, 0)
			}
			return nil
		},
	}

	cmd.Flags().BoolVarP(&printReceipt, "print-receipt", "p", false, "print tx receipt")
	waspcmd.WithWaspNodeFlag(cmd, &node)
	withChainFlag(cmd, &chain)

	return cmd
}
