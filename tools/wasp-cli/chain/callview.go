package chain

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/clients"
	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/tools/wasp-cli/util"
)

func initCallViewCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "call-view <name> <funcname> [params]",
		Short: "Call a contract view function",
		Long:  "Call contract <name>, view function <funcname> with given params.",
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			client := cliclients.WaspClientForIndex()

			contractName := args[0]
			funcName := args[1]
			params := util.EncodeParams(args[2:])

			result, _, err := client.RequestsApi.CallView(context.Background()).ContractCallViewRequest(apiclient.ContractCallViewRequest{
				ChainId:      GetCurrentChainID().String(),
				ContractName: contractName,
				FunctionName: funcName,
				Arguments:    clients.JSONDictToAPIJSONDict(params.JSONDict()),
			}).Execute()
			log.Check(err)

			decodedResult, err := clients.APIJsonDictToDict(*result)
			log.Check(err)

			util.PrintDictAsJSON(decodedResult)
		},
	}
}
