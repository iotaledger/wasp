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
)

func postRequest(hname, fname string, params chainclient.PostRequestParams, offLedger, adjustStorageDeposit bool) {
	apiClient := cliclients.WaspClientForIndex()
	scClient := cliclients.SCClient(apiClient, isc.Hn(hname))

	if offLedger {
		params.Nonce = uint64(time.Now().UnixNano())
		util.WithOffLedgerRequest(config.GetCurrentChainID(), func() (isc.OffLedgerRequest, error) {
			return scClient.PostOffLedgerRequest(fname, params)
		})
		return
	}

	if !adjustStorageDeposit {
		// check if there are enough funds for SD
		output := transaction.MakeBasicOutput(
			config.GetCurrentChainID().AsAddress(),
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
			true,
		)
		util.SDAdjustmentPrompt(output)
	}

	util.WithSCTransaction(config.GetCurrentChainID(), func() (*iotago.Transaction, error) {
		return scClient.PostRequest(fname, params)
	})
}

func initPostRequestCmd() *cobra.Command {
	var transfer []string
	var allowance []string
	var offLedger bool
	var adjustStorageDeposit bool

	cmd := &cobra.Command{
		Use:   "post-request <name> <funcname> [params]",
		Short: "Post a request to a contract",
		Long:  "Post a request to contract <name>, function <funcname> with given params.",
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			hname := args[0]
			fname := args[1]

			allowanceTokens := util.ParseFungibleTokens(allowance)
			params := chainclient.PostRequestParams{
				Args:      util.EncodeParams(args[2:]),
				Transfer:  util.ParseFungibleTokens(transfer),
				Allowance: allowanceTokens,
			}
			postRequest(hname, fname, params, offLedger, adjustStorageDeposit)
		},
	}

	cmd.Flags().StringSliceVarP(&allowance, "allowance", "l", []string{},
		"include allowance as part of the transaction. Format: <token-id>:<amount>,<token-id>:amount...")

	cmd.Flags().StringSliceVarP(&transfer, "transfer", "t", []string{},
		"include a funds transfer as part of the transaction. Format: <token-id>:<amount>,<token-id>:amount...",
	)
	cmd.Flags().BoolVarP(&offLedger, "off-ledger", "o", false,
		"post an off-ledger request",
	)
	cmd.Flags().BoolVarP(&adjustStorageDeposit, "adjust-storage-deposit", "s", false, "adjusts the amount of base tokens sent, if it's lower than the min storage deposit required")

	return cmd
}
