package chain

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/v2/clients/apiextensions"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/util"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/waspcmd"
)

func initCallViewCmd() *cobra.Command {
	var node string
	var chain string

	cmd := &cobra.Command{
		Use:   "call-view <name> <funcname> [params]",
		Short: "Call a contract view function",
		Long:  "Call contract <name>, view function <funcname> with given params.",
		Args:  cobra.MinimumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			node = waspcmd.DefaultWaspNodeFallback(node)
			chain = defaultChainFallback(chain)
			ctx := context.Background()
			client := cliclients.WaspClientWithVersionCheck(ctx, node)

			contractName := args[0]
			funcName := args[1]
			params := util.EncodeParams(args[2:])

			msg := isc.NewMessage(isc.Hn(contractName), isc.Hn(funcName), params)

			result, _, err := client.ChainsAPI.CallView(ctx).
				ContractCallViewRequest(apiextensions.CallViewReq(msg)).Execute() //nolint:bodyclose // false positive

			log.Check(err)

			decodedResult, err := apiextensions.APIResultToCallArgs(result)
			log.Check(err)

			util.PrintCallResultsAsJSON(decodedResult)
		},
	}
	waspcmd.WithWaspNodeFlag(cmd, &node)
	withChainFlag(cmd, &chain)

	return cmd
}
