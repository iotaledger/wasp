//go:build exclude

// Excluded for now as we right now don't deploying new contracts
package chain

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/tools/wasp-cli/util"
	"github.com/iotaledger/wasp/tools/wasp-cli/waspcmd"
)

func initDeployContractCmd() *cobra.Command {
	var node string
	var chain string

	cmd := &cobra.Command{
		Use:   "deploy-contract <vmtype> <name> <filename|program-hash> [init-params]",
		Short: "Deploy a contract in the chain",
		Args:  cobra.MinimumNArgs(4),
		Run: func(cmd *cobra.Command, args []string) {
			node = waspcmd.DefaultWaspNodeFallback(node)
			chain = defaultChainFallback(chain)

			chainID := config.GetChain(chain)
			client := cliclients.WaspClient(node)
			vmtype := args[0]
			name := args[1]
			initParams := util.EncodeParams(args[3:], chainID)

			var progHash hashing.HashValue

			switch vmtype {
			case vmtypes.Core:
				log.Fatal("cannot manually deploy core contracts")

			case vmtypes.Native:
				var err error
				progHash, err = hashing.HashValueFromHex(args[2])
				log.Check(err)

			default:
				log.Fatal("you can only deploy native (non-core) contracts")

			}
			deployContract(client, chainID, node, name, progHash, initParams)
		},
	}
	waspcmd.WithWaspNodeFlag(cmd, &node)
	withChainFlag(cmd, &chain)
	return cmd
}

func deployContract(client *apiclient.APIClient, chainID isc.ChainID, node, name string, progHash hashing.HashValue, initParams isc.CallArguments) {
	util.WithOffLedgerRequest(chainID, node, func() (isc.OffLedgerRequest, error) {
		return cliclients.ChainClient(client, chainID).PostOffLedgerRequest(context.Background(),
			root.FuncDeployContract.Message(name, progHash, initParams),
		)
	})
}
