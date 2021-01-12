package chain

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

func listContractsCmd(args []string) {
	info, err := SCClient(root.Interface.Hname()).CallView(root.FuncGetChainInfo, nil)
	log.Check(err)

	contracts, err := root.DecodeContractRegistry(collections.NewMapReadOnly(info, root.VarContractRegistry))
	log.Check(err)

	feeColor, defaultOwnerFee, defaultValidatorFee, err := root.GetDefaultFeeInfo(info)
	log.Check(err)

	log.Printf("Total %d contracts in chain %s\n", len(contracts), GetCurrentChainID())

	header := []string{
		"hname",
		"name",
		"description",
		"proghash",
		"creator",
		"owner fee",
		"validator fee",
	}
	rows := make([][]string, len(contracts))
	i := 0
	for hname, c := range contracts {
		creator := ""
		if c.HasCreator() {
			creator = c.Creator.String()
		}

		ownerFee := c.OwnerFee
		if ownerFee == 0 {
			ownerFee = defaultOwnerFee
		}
		validatorFee := c.ValidatorFee
		if validatorFee == 0 {
			validatorFee = defaultValidatorFee
		}

		rows[i] = []string{
			hname.String(),
			c.Name,
			c.Description,
			c.ProgramHash.String(),
			creator,
			fmt.Sprintf("%d %s", ownerFee, feeColor),
			fmt.Sprintf("%d %s", validatorFee, feeColor),
		}
		i++
	}
	log.PrintTable(header, rows)
}
