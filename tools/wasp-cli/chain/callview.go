package chain

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/clients/apiextensions"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/tools/wasp-cli/util"
	"github.com/iotaledger/wasp/tools/wasp-cli/waspcmd"
)

func initCallViewCmd() *cobra.Command {
	var node string
	cmd := &cobra.Command{
		Use:   "call-view <name> <funcname> [params]",
		Short: "Call a contract view function",
		Long:  "Call contract <name>, view function <funcname> with given params.",
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			node = waspcmd.DefaultWaspNodeFallback(node)
			client := cliclients.WaspClient(node)

			contractName := args[0]
			funcName := args[1]
			params := util.EncodeParams(args[2:])

			result, _, err := client.RequestsApi.CallView(context.Background()).ContractCallViewRequest(apiclient.ContractCallViewRequest{
				ChainId:      config.GetCurrentChainID().String(),
				ContractName: contractName,
				FunctionName: funcName,
				Arguments:    apiextensions.JSONDictToAPIJSONDict(params.JSONDict()),
			}).Execute()
			log.Check(err)

			decodedResult, err := apiextensions.APIJsonDictToDict(*result)
			log.Check(err)

			util.PrintDictAsJSON(decodedResult)
		},
	}
	waspcmd.WithWaspNodeFlag(cmd, &node)
	return cmd
}
