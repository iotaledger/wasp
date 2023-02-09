package chain

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/tools/wasp-cli/waspcmd"
)

func initEventsCmd() *cobra.Command {
	var node string
	cmd := &cobra.Command{
		Use:   "events <name>",
		Short: "Show events of contract <name>",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			node = waspcmd.DefaultSingleNodeFallback(node)
			client := cliclients.WaspClient(node)
			contractHName := isc.Hn(args[0]).String()

			events, _, err := client.CorecontractsApi.
				BlocklogGetEventsOfContract(context.Background(), config.GetCurrentChainID().String(), contractHName).
				Execute()

			log.Check(err)
			logEvents(events)
		},
	}
	waspcmd.WithSingleWaspNodesFlag(cmd, &node)
	return cmd
}
