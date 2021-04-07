package chain

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/requestargs"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/tools/wasp-cli/util"
	"github.com/spf13/cobra"
)

var deployContractCmd = &cobra.Command{
	Use:   "deploy-contract <vmtype> <name> <description> <filename>",
	Short: "Deploy a contract in the chain",
	Args:  cobra.ExactArgs(4),
	Run: func(cmd *cobra.Command, args []string) {
		vmtype := args[0]
		name := args[1]
		description := args[2]
		filename := args[3]

		blobFieldValues := codec.MakeDict(map[string]interface{}{
			blob.VarFieldVMType:             vmtype,
			blob.VarFieldProgramDescription: description,
			blob.VarFieldProgramBinary:      util.ReadFile(filename),
		})

		progHash := uploadBlob(blobFieldValues, true)

		util.WithSCTransaction(GetCurrentChainID(), func() (*ledgerstate.Transaction, error) {
			return Client().PostRequest(
				root.Interface.Hname(),
				coretypes.Hn(root.FuncDeployContract),
				chainclient.PostRequestParams{
					Args: requestargs.New().AddEncodeSimpleMany(codec.MakeDict(map[string]interface{}{
						root.ParamName:        name,
						root.ParamDescription: description,
						root.ParamProgramHash: progHash,
					})),
				},
			)
		})
	},
}
