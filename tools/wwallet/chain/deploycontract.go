package chain

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/root"
	"github.com/iotaledger/wasp/tools/wwallet/util"
)

func deployContractCmd(args []string) {
	if len(args) != 4 {
		check(fmt.Errorf("Usage: %s chain deploy-contract <vmtype> <name> <description> <filename>", os.Args[0]))
	}

	vmtype := args[0]
	name := args[1]
	description := args[2]
	filename := args[3]

	util.WithSCTransaction(func() (*sctransaction.Transaction, error) {
		return Client().PostRequest(
			root.Hname,
			coretypes.Hn(root.FuncDeployContract),
			nil,
			nil,
			map[string]interface{}{
				root.ParamName:          name,
				root.ParamVMType:        vmtype,
				root.ParamDescription:   description,
				root.ParamProgramBinary: readBinary(filename),
			},
		)
	})
}

func readBinary(fname string) []byte {
	b, err := ioutil.ReadFile(fname)
	check(err)
	return b
}
