package chain

import (
	"strconv"
	"strings"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/colored"
	"github.com/iotaledger/wasp/packages/iscp/request"
	"github.com/iotaledger/wasp/packages/iscp/requestargs"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/tools/wasp-cli/util"
	"github.com/spf13/cobra"
)

func postRequestCmd() *cobra.Command {
	var transfer []string
	var offLedger bool

	cmd := &cobra.Command{
		Use:   "post-request <name> <funcname> [params]",
		Short: "Post a request to a contract",
		Long:  "Post a request to contract <name>, function <funcname> with given params.",
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			fname := args[1]
			params := chainclient.PostRequestParams{
				Args:     requestargs.New().AddEncodeSimpleMany(util.EncodeParams(args[2:])),
				Transfer: parseColoredBalances(transfer),
			}

			scClient := SCClient(iscp.Hn(args[0]))

			if offLedger {
				params.Nonce = uint64(time.Now().UnixNano())
				util.WithOffLedgerRequest(GetCurrentChainID(), func() (*request.OffLedger, error) {
					return scClient.PostOffLedgerRequest(fname, params)
				})
			} else {
				util.WithSCTransaction(GetCurrentChainID(), func() (*ledgerstate.Transaction, error) {
					return scClient.PostRequest(fname, params)
				})
			}
		},
	}

	cmd.Flags().StringSliceVarP(&transfer, "transfer", "t", []string{"IOTA:1"},
		"include a funds transfer as part of the transaction. Format: <color>:<amount>,<color>:amount...",
	)
	cmd.Flags().BoolVarP(&offLedger, "off-ledger", "o", false,
		"post an off-ledger request",
	)

	return cmd
}

func colorFromString(s string) colored.Color {
	if s == colored.IOTA.String() {
		return colored.IOTA
	}
	c, err := colored.ColorFromBase58EncodedString(s)
	log.Check(err)
	return c
}

func parseColoredBalances(args []string) colored.Balances {
	cb := colored.NewBalances()
	for _, tr := range args {
		parts := strings.Split(tr, ":")
		if len(parts) != 2 {
			log.Fatalf("colored balances syntax: <color>:<amount>,<color:amount>... -- Example: IOTA:100")
		}
		col := colorFromString(parts[0])
		amount, err := strconv.Atoi(parts[1])
		log.Check(err)
		cb.Set(col, uint64(amount))
	}
	return cb
}
