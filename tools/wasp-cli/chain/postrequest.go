package chain

import (
	"time"

	"github.com/spf13/cobra"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/clients/chainclient"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/vm/gas"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/util"
	"github.com/iotaledger/wasp/tools/wasp-cli/waspcmd"
)

func postRequest(nodeName, chain, hname, fname string, params chainclient.PostRequestParams, offLedger, adjustStorageDeposit bool) {
	chainID := config.GetChain(chain)

	apiClient := cliclients.WaspClient(nodeName)
	scClient := cliclients.SCClient(apiClient, chainID, isc.Hn(hname))

	if offLedger {
		params.Nonce = uint64(time.Now().UnixNano())
		util.WithOffLedgerRequest(chainID, nodeName, func() (isc.OffLedgerRequest, error) {
			return scClient.PostOffLedgerRequest(fname, params)
		})
		return
	}

	if !adjustStorageDeposit {
		// check if there are enough funds for SD
		output := transaction.MakeBasicOutput(
			chainID.AsAddress(),
			scClient.ChainClient.KeyPair.Address(),
			params.Transfer,
			&isc.RequestMetadata{
				SenderContract: 0,
				TargetContract: isc.Hn(hname),
				EntryPoint:     isc.Hn(fname),
				Params:         params.Args,
				Allowance:      params.Allowance,
				GasBudget:      gas.MaxGasPerRequest,
			},
			isc.SendOptions{},
		)
		util.SDAdjustmentPrompt(output)
	}

	util.WithSCTransaction(config.GetChain(chain), nodeName, func() (*iotago.Transaction, error) {
		return scClient.PostRequest(fname, params)
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
			hname := args[0]
			fname := args[1]

			allowanceTokens := util.ParseFungibleTokens(postRequestParams.allowance)
			params := chainclient.PostRequestParams{
				Args:      util.EncodeParams(args[2:]),
				Transfer:  util.ParseFungibleTokens(postRequestParams.transfer),
				Allowance: allowanceTokens,
			}
			postRequest(node, chain, hname, fname, params, postRequestParams.offLedger, postRequestParams.adjustStorageDeposit)
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
