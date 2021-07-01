package chain

import (
	"fmt"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/tools/wasp-cli/util"
)

var listAccountsCmd = &cobra.Command{
	Use:   "list-accounts",
	Short: "List accounts in chain",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		ret, err := SCClient(accounts.Interface.Hname()).CallView(accounts.FuncViewAccounts)
		log.Check(err)

		log.Printf("Total %d account(s) in chain %s\n", len(ret), GetCurrentChainID().Base58())

		header := []string{"agentid"}
		rows := make([][]string, len(ret))
		i := 0
		for k := range ret {
			agentID, _, err := codec.DecodeAgentID([]byte(k))
			log.Check(err)
			rows[i] = []string{agentID.String()}
			i++
		}
		log.PrintTable(header, rows)
	},
}

var balanceCmd = &cobra.Command{
	Use:   "balance <agentid>",
	Short: "Show balance of on-chain account",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		agentID, err := coretypes.NewAgentIDFromString(args[0])
		log.Check(err)

		ret, err := SCClient(accounts.Interface.Hname()).CallView(accounts.FuncViewBalance,
			dict.Dict{
				accounts.ParamAgentID: agentID.Bytes(),
			})
		log.Check(err)

		header := []string{"color", "amount"}
		rows := make([][]string, len(ret))
		i := 0
		for k, v := range ret {
			color, _, err := ledgerstate.ColorFromBytes([]byte(k))
			log.Check(err)
			bal, _, err := codec.DecodeUint64(v)
			log.Check(err)

			rows[i] = []string{color.String(), fmt.Sprintf("%d", bal)}
			i++
		}
		log.PrintTable(header, rows)
	},
}

var depositCmd = &cobra.Command{
	Use:   "deposit <color>:<amount> [<color>:amount ...]",
	Short: "Deposit funds into sender's on-chain account",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		util.WithSCTransaction(GetCurrentChainID(), func() (*ledgerstate.Transaction, error) {
			return SCClient(accounts.Interface.Hname()).PostRequest(
				accounts.FuncDeposit,
				chainclient.PostRequestParams{
					Transfer: parseColoredBalances(args),
				},
			)
		})
	},
}
