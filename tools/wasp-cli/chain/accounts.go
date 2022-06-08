package chain

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/tools/wasp-cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/tools/wasp-cli/util"
	"github.com/spf13/cobra"
)

const iotaTokenStr = "iota"

var listAccountsCmd = &cobra.Command{
	Use:   "list-accounts",
	Short: "List accounts in chain",
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
	Use:   "balance <agentid>",
	Short: "Show balance of on-chain account",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if parameters.L1 == nil {
			config.L1Client() // this will fill parameters.L1 with data from the L1 node
		}
		agentID, err := iscp.NewAgentIDFromString(args[0])
		log.Check(err)

		ret, err := SCClient(accounts.Contract.Hname()).CallView(accounts.ViewBalance.Name, dict.Dict{
			accounts.ParamAgentID: agentID.Bytes(),
		})
		log.Check(err)

		header := []string{"token", "amount"}
		rows := make([][]string, len(ret))
		i := 0
		for k, v := range ret {
			tokenStr := iotaTokenStr
			if !iscp.IsIota([]byte(k)) {
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

var depositCmd = &cobra.Command{
	Use:   "deposit <color>:<amount> [<color>:amount ...]",
	Short: "Deposit funds into sender's on-chain account",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		util.WithSCTransaction(GetCurrentChainID(), func() (*iotago.Transaction, error) {
			return SCClient(accounts.Contract.Hname()).PostRequest(
				accounts.FuncDeposit.Name,
				chainclient.PostRequestParams{
					Transfer: parseAssets(args),
				},
			)
		})
	},
}
