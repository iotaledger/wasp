package chain

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/kv/datatypes"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/root"
)

func listContractsCmd(args []string) {
	ret, err := SCClient(root.Interface.Hname()).CallView(root.FuncGetInfo, nil)
	check(err)
	contracts, err := root.DecodeContractRegistry(datatypes.NewMustMap(ret, root.VarContractRegistry))
	check(err)
	for hname, c := range contracts {
		fmt.Printf("[hname: %s]: [Name: %s] [Description: %s] [proghash: %s]\n", hname, c.Name, c.Description, c.ProgramHash.String())
	}
}
