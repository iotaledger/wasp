package chain

import (
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/tools/wasp-cli/util"
	"github.com/spf13/cobra"
)

func postRequestCmd() *cobra.Command {
	var transfer []string
	var allowance []string
	var offLedger bool

	cmd := &cobra.Command{
		Use:   "post-request <name> <funcname> [params]",
		Short: "Post a request to a contract",
		Long:  "Post a request to contract <name>, function <funcname> with given params.",
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			fname := args[1]

			allowanceTokens := util.ParseFungibleTokens(allowance)
			params := chainclient.PostRequestParams{
				Args:      util.EncodeParams(args[2:]),
				Transfer:  util.ParseFungibleTokens(transfer),
				Allowance: iscp.NewAllowanceFungibleTokens(allowanceTokens),
			}

			scClient := SCClient(iscp.Hn(args[0]))

			if offLedger {
				params.Nonce = uint64(time.Now().UnixNano())
				util.WithOffLedgerRequest(GetCurrentChainID(), func() (iscp.OffLedgerRequest, error) {
					return scClient.PostOffLedgerRequest(fname, params)
				})
			} else {
				util.WithSCTransaction(GetCurrentChainID(), func() (*iotago.Transaction, error) {
					return scClient.PostRequest(fname, params)
				})
			}
		},
	}

	cmd.Flags().StringSliceVarP(&allowance, "allowance", "l", []string{},
		"include allowances as part of the transaction. Format: <token-id>:<amount>,<token-id>:amount...")

	cmd.Flags().StringSliceVarP(&transfer, "transfer", "t", []string{},
		"include a funds transfer as part of the transaction. Format: <token-id>:<amount>,<token-id>:amount...",
	)
	cmd.Flags().BoolVarP(&offLedger, "off-ledger", "o", false,
		"post an off-ledger request",
	)

	return cmd
}
