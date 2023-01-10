package chain

import (
	"math/big"
	"strings"

	"github.com/spf13/cobra"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/gas"
	"github.com/iotaledger/wasp/tools/wasp-cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/tools/wasp-cli/util"
	"github.com/iotaledger/wasp/tools/wasp-cli/wallet"
)

var listAccountsCmd = &cobra.Command{
	Use:   "list-accounts",
	Short: "List L2 accounts",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		ret, err := SCClient(accounts.Contract.Hname()).CallView(accounts.ViewAccounts.Name, nil)
		log.Check(err)

		log.Printf("Total %d account(s) in chain %s\n", len(ret), GetCurrentChainID().String())

		header := []string{"agentid"}
		rows := make([][]string, len(ret))
		i := 0
		for k := range ret {
			agentID, err := codec.DecodeAgentID([]byte(k))
			log.Check(err)
			rows[i] = []string{agentID.String()}
			i++
		}
		log.PrintTable(header, rows)
	},
}

var balanceCmd = &cobra.Command{
	Use:   "balance [<agentid>]",
	Short: "Show the L2 balance of the given account",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		var agentID isc.AgentID
		if len(args) == 0 {
			agentID = isc.NewAgentID(wallet.Load().Address())
		} else {
			var err error
			agentID, err = isc.NewAgentIDFromString(args[0])
			log.Check(err)
		}

		ret, err := SCClient(accounts.Contract.Hname()).CallView(accounts.ViewBalance.Name, dict.Dict{
			accounts.ParamAgentID: agentID.Bytes(),
		})
		log.Check(err)

		header := []string{"token", "amount"}
		rows := make([][]string, len(ret))
		i := 0
		for k, v := range ret {
			tokenStr := util.BaseTokenStr
			if !isc.IsBaseToken([]byte(k)) {
				tokenStr = codec.MustDecodeNativeTokenID([]byte(k)).String()
			}
			bal, err := codec.DecodeBigIntAbs(v)
			log.Check(err)

			rows[i] = []string{tokenStr, bal.String()}
			i++
		}
		log.PrintTable(header, rows)
	},
}

func getTokensForRequestFee() *isc.Allowance {
	apiAddress := config.WaspAPI()
	if apiAddress == "" {
		// no wasp api defined, assume default configuration
		log.Printf("wasp webapi not defined, using default values to calculate the fees")
		baseTokens := gas.MinGasPerRequest / gas.DefaultGasFeePolicy().GasPerToken // default value
		return isc.NewAllowanceBaseTokens(baseTokens)
	}
	client := config.WaspClient(apiAddress)
	gasPolicy, err := client.GetGasFeePolicy(GetCurrentChainID())
	log.Check(err)
	amount := gas.MinGasPerRequest / gasPolicy.GasPerToken
	if isc.IsEmptyNativeTokenID(gasPolicy.GasFeeTokenID) || isc.IsBaseToken(gasPolicy.GasFeeTokenID[:]) {
		return isc.NewAllowanceBaseTokens(amount)
	}

	return isc.NewAllowanceFungibleTokens(isc.NewFungibleTokens(0, iotago.NativeTokens{{
		ID:     gasPolicy.GasFeeTokenID,
		Amount: new(big.Int).SetUint64(amount),
	}}))
}

func depositCmd() *cobra.Command {
	var adjustStorageDeposit bool

	cmd := &cobra.Command{
		Use:   "deposit [<agentid>] <token-id>:<amount>, [<token-id>:amount ...]",
		Short: "Deposit L1 funds into the given (default: your) L2 account",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if strings.Contains(args[0], ":") {
				// deposit to own agentID
				tokens := util.ParseFungibleTokens(args)
				util.WithSCTransaction(GetCurrentChainID(), func() (*iotago.Transaction, error) {
					return SCClient(accounts.Contract.Hname()).PostRequest(
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
				allowance := isc.NewAllowanceFungibleTokens(tokens.Clone())

				util.WithSCTransaction(GetCurrentChainID(), func() (*iotago.Transaction, error) {
					return SCClient(accounts.Contract.Hname()).PostRequest(
						accounts.FuncTransferAllowanceTo.Name,
						chainclient.PostRequestParams{
							Args: dict.Dict{
								accounts.ParamAgentID:          agentID.Bytes(),
								accounts.ParamForceOpenAccount: codec.EncodeBool(true),
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

	return cmd
}
