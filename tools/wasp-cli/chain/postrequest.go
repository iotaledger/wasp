package chain

import (
	"context"
	"time"

	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/clients/chainclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/util"
	"github.com/iotaledger/wasp/tools/wasp-cli/waspcmd"
)

//nolint:unparam
func postRequest(ctx context.Context, client *apiclient.APIClient, chain string, msg isc.Message, params chainclient.PostRequestParams, offLedger, adjustStorageDeposit bool) {
	chainID := config.GetChain(chain)
	chainClient := cliclients.ChainClient(client, chainID)

	if offLedger {
		util.WithOffLedgerRequest(ctx, client, func() (isc.OffLedgerRequest, error) {
			return chainClient.PostOffLedgerRequest(ctx, msg, params)
		})
		return
	}

	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	util.WithSCTransaction(ctx, client, func() (*iotajsonrpc.IotaTransactionBlockResponse, error) {
		return chainClient.PostRequest(ctx, msg, params)
	})
}

func initPostRequestCmd() *cobra.Command {
	var (
		postRequestParams postRequestParams
		node              string
		chain             string
	)

	cmd := &cobra.Command{
		Use:   "post-request <name> <funcname> [params]",
		Short: "Post a request to a contract",
		Long:  "Post a request to contract <name>, function <funcname> with given params.",
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			node = waspcmd.DefaultWaspNodeFallback(node)
			chain = defaultChainFallback(chain)
			ctx := context.Background()
			client := cliclients.WaspClientWithVersionCheck(ctx, node)

			cname := args[0]
			fname := args[1]
			params := util.EncodeParams(args[2:])
			msg := isc.NewMessage(isc.Hn(cname), isc.Hn(fname), params)

			// allowanceTokens := util.ParseFungibleTokens(postRequestParams.allowance)
			postParams := chainclient.PostRequestParams{
				Transfer:  isc.NewAssets(100000000),
				Allowance: isc.NewAssets(1000000),
			}
			postRequest(ctx, client, chain, msg, postParams, postRequestParams.offLedger, postRequestParams.adjustStorageDeposit)
		},
	}

	waspcmd.WithWaspNodeFlag(cmd, &node)
	withChainFlag(cmd, &chain)
	postRequestParams.initFlags(cmd)

	return cmd
}

type postRequestParams struct {
	transfer             []string
	allowance            []string
	offLedger            bool
	adjustStorageDeposit bool
}

func (p *postRequestParams) initFlags(cmd *cobra.Command) {
	cmd.Flags().StringSliceVarP(&p.allowance, "allowance", "l", []string{},
		"include allowance as part of the transaction. Format: <token-id>:<amount>,<token-id>:amount...")

	cmd.Flags().StringSliceVarP(&p.transfer, "transfer", "t", []string{},
		"include a funds transfer as part of the transaction. Format: <token-id>:<amount>,<token-id>:amount...",
	)
	cmd.Flags().BoolVarP(&p.offLedger, "off-ledger", "o", false,
		"post an off-ledger request",
	)
	cmd.Flags().BoolVarP(&p.adjustStorageDeposit, "adjust-storage-deposit", "s", false, "adjusts the amount of base tokens sent, if it's lower than the min storage deposit required")
}
