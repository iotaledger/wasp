package chain

import (
	"strings"

	"github.com/spf13/cobra"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/tools/wasp-cli/util"
	"github.com/iotaledger/wasp/tools/wasp-cli/wallet"
)

func initListAccountsCmd() *cobra.Command {
	return &cobra.Command{
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
}

func initBalanceCmd() *cobra.Command {
	return &cobra.Command{
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
}

func initDepositCmd() *cobra.Command {
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

	return cmd
}
