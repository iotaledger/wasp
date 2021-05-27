package chain

import (
	"strconv"
	"strings"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/requestargs"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/tools/wasp-cli/util"
	"github.com/spf13/cobra"
)

func postRequestCmd() *cobra.Command {
	var transfer []string

	cmd := &cobra.Command{
		Use:   "post-request <name> <funcname> [params]",
		Short: "Post a request to a contract",
		Long:  "Post a request to contract <name>, function <funcname> with given params.",
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			cb := make(map[ledgerstate.Color]uint64)
			for _, tr := range transfer {
				parts := strings.Split(tr, ":")
				if len(parts) != 2 {
					log.Fatal("Syntax for --transfer: <color>:<amount>,<color:amount>...\nExample: IOTA:100")
				}
				color := colorFromString(parts[0])
				amount, err := strconv.Atoi(parts[1])
				log.Check(err)
				cb[color] = uint64(amount)
			}

			util.WithSCTransaction(GetCurrentChainID(), func() (*ledgerstate.Transaction, error) {
				return SCClient(coretypes.Hn(args[0])).PostRequest(
					args[1],
					chainclient.PostRequestParams{
						Args:     requestargs.New().AddEncodeSimpleMany(util.EncodeParams(args[2:])),
						Transfer: ledgerstate.NewColoredBalances(cb),
					},
				)
			})
		},
	}

	cmd.Flags().StringSliceVarP(&transfer, "transfer", "", []string{"IOTA:1"},
		"include a funds transfer as part of the transaction. Format: <color>:<amount>,<color>:amount...",
	)

	return cmd
}

func colorFromString(s string) ledgerstate.Color {
	if s == ledgerstate.ColorIOTA.String() {
		return ledgerstate.ColorIOTA
	}
	c, err := ledgerstate.ColorFromBase58EncodedString(s)
	log.Check(err)
	return c
}
