package chain

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/clients/chainclient"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
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
	cmd := &cobra.Command{
		Use:   "deploy-contract <vmtype> <name> <description> <filename|program-hash> [init-params]",
		Short: "Deploy a contract in the chain",
		Args:  cobra.MinimumNArgs(4),
		Run: func(cmd *cobra.Command, args []string) {
			node = waspcmd.DefaultWaspNodeFallback(node)
			client := cliclients.WaspClient(node)
			vmtype := args[0]
			name := args[1]
			description := args[2]
			initParams := util.EncodeParams(args[4:])

			var progHash hashing.HashValue

			switch vmtype {
			case vmtypes.Core:
				log.Fatal("cannot manually deploy core contracts")

			case vmtypes.Native:
				var err error
				progHash, err = hashing.HashValueFromHex(args[3])
				log.Check(err)

			default:
				filename := args[3]
				blobFieldValues := codec.MakeDict(map[string]interface{}{
					blob.VarFieldVMType:             vmtype,
					blob.VarFieldProgramDescription: description,
					blob.VarFieldProgramBinary:      util.ReadFile(filename),
				})
				progHash = uploadBlob(client, blobFieldValues)
			}
			deployContract(client, node, name, description, progHash, initParams)
		},
	}
	waspcmd.WithWaspNodeFlag(cmd, &node)
	return cmd
}

func deployContract(client *apiclient.APIClient, node, name, description string, progHash hashing.HashValue, initParams dict.Dict) {
	util.WithOffLedgerRequest(config.GetCurrentChainID(), node, func() (isc.OffLedgerRequest, error) {
		args := codec.MakeDict(map[string]interface{}{
			root.ParamName:        name,
			root.ParamDescription: description,
			root.ParamProgramHash: progHash,
		})
		args.Extend(initParams)
		return cliclients.ChainClient(client).PostOffLedgerRequest(context.Background(),
			root.Contract.Hname(),
			root.FuncDeployContract.Hname(),
			chainclient.PostRequestParams{
				Args: args,
			},
		)
	})
}
