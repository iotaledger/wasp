package chain

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/root"
)

func deployContractCmd(args []string) {
	if len(args) != 4 {
		check(fmt.Errorf("Usage: %s chain deploy-contract <vmtype> <name> <description> <filename>", os.Args[0]))
	}

	vmtype := args[0]
	name := args[1]
	description := args[2]
	filename := args[3]

	tx, err := Client().PostRequest(
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
	check(err)
	err = Client().WaspClient.WaitUntilAllRequestsProcessed(tx, 1*time.Minute)
	check(err)
}

func readBinary(fname string) []byte {
	b, err := ioutil.ReadFile(fname)
	check(err)
	return b
}
