package chain

import (
	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/vm/vmtypes"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/tools/wasp-cli/util"
	"github.com/spf13/cobra"
)

var deployContractCmd = &cobra.Command{
	Use:   "deploy-contract <vmtype> <name> <description> <filename|program-hash> [init-params]",
	Short: "Deploy a contract in the chain",
	Args:  cobra.MinimumNArgs(4),
	Run: func(cmd *cobra.Command, args []string) {
		vmtype := args[0]
		name := args[1]
		description := args[2]
		initParams := util.EncodeParams(args[4:])

		var progHash hashing.HashValue

		switch vmtype {
		case vmtypes.Core:
			log.Fatalf("cannot manually deploy core contracts")

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
			progHash = uploadBlob(blobFieldValues)
		}

		deployContract(name, description, progHash, initParams)
	},
}

func deployContract(name, description string, progHash hashing.HashValue, initParams dict.Dict) {
	util.WithOffLedgerRequest(GetCurrentChainID(), func() (iscp.OffLedgerRequest, error) {
		args := codec.MakeDict(map[string]interface{}{
			root.ParamName:        name,
			root.ParamDescription: description,
			root.ParamProgramHash: progHash,
		})
		args.Extend(initParams)
		return Client().PostOffLedgerRequest(
			root.Contract.Hname(),
			root.FuncDeployContract.Hname(),
			chainclient.PostRequestParams{
				Args: args,
			},
		)
	})
}
