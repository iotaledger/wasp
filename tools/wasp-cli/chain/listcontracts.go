package chain

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/spf13/cobra"
)

var listContractsCmd = &cobra.Command{
	Use:   "list-contracts",
	Short: "List deployed contracts in chain",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		records, err := SCClient(root.Contract.Hname()).CallView(root.FuncGetContractRecords.Name, nil)
		log.Check(err)
		contracts, err := root.DecodeContractRegistry(collections.NewMapReadOnly(records, root.VarContractRegistry))
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

			fees, err := SCClient(governance.Contract.Hname()).CallView(governance.FuncGetFeeInfo.Name, dict.Dict{
				governance.ParamHname: c.Hname().Bytes(),
			})
			log.Check(err)

			ownerFee, _, err := codec.DecodeUint64(fees.MustGet(governance.VarOwnerFee))
			log.Check(err)

			validatorFee, _, err := codec.DecodeUint64(fees.MustGet(governance.VarValidatorFee))
			log.Check(err)

			feeColor, _, err := codec.DecodeColor(fees.MustGet(governance.VarFeeColor))
			log.Check(err)

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
	},
}
