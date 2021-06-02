package chain

import (
	"fmt"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

var listAccountsCmd = &cobra.Command{
	Use:   "list-accounts",
	Short: "List accounts in chain",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		ret, err := SCClient(accounts.Interface.Hname()).CallView(accounts.FuncAccounts)
		log.Check(err)

		log.Printf("Total %d account(s) in chain %s\n", len(ret), GetCurrentChainID().Base58())

		header := []string{"agentid"}
		rows := make([][]string, len(ret))
		i := 0
		for k := range ret {
			agentId, _, err := codec.DecodeAgentID([]byte(k))
			log.Check(err)
			rows[i] = []string{agentId.String()}
			i++
		}
		log.PrintTable(header, rows)
	},
}

var balanceCmd = &cobra.Command{
	Use:   "balance <agentid>",
	Short: "Show balance of account in chain",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		agentID, err := coretypes.NewAgentIDFromString(args[0])
		log.Check(err)

		ret, err := SCClient(accounts.Interface.Hname()).CallView(accounts.FuncBalance,
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
			bal, err := util.Uint64From8Bytes(v)
			log.Check(err)

			rows[i] = []string{color.String(), fmt.Sprintf("%d", bal)}
			i++
		}
		log.PrintTable(header, rows)
	},
}
