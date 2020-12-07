package chain

import (
	"github.com/iotaledger/wasp/packages/kv/datatypes"
	"github.com/iotaledger/wasp/packages/vm/builtinvm/root"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

func listContractsCmd(args []string) {
	ret, err := SCClient(root.Interface.Hname()).CallView(root.FuncGetInfo, nil)
	log.Check(err)
	contracts, err := root.DecodeContractRegistry(datatypes.NewMustMap(ret, root.VarContractRegistry))
	log.Check(err)

	log.Printf("Total %d contracts in chain %s\n", len(contracts), GetCurrentChainID())

	header := []string{"hname", "name", "description", "proghash"}
	rows := make([][]string, len(contracts))
	i := 0
	for hname, c := range contracts {
		rows[i] = []string{
			hname.String(),
			c.Name,
			c.Description,
			c.ProgramHash.String(),
		}
		i++
	}
	log.PrintTable(header, rows)
}
