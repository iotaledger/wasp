package chain

import (
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/tools/wasp-cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Show information about the chain",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		chain, err := config.WaspClient().GetChainRecord(GetCurrentChainID())
		log.Check(err)

		committee, err := config.WaspClient().GetCommitteeForChain(chain.ChainID)
		log.Check(err)

		log.Printf("Chain ID: %s\n", chain.ChainID.Base58())
		log.Printf("Committee nodes: %+v\n", committee.Nodes)
		log.Printf("Active: %v\n", chain.Active)

		if chain.Active {
			info, err := SCClient(governance.Contract.Hname()).CallView(governance.FuncGetChainInfo.Name, nil)
			log.Check(err)

			description, err := codec.DecodeString(info.MustGet(governance.VarDescription), "")
			log.Check(err)
			log.Printf("Description: %s\n", description)

			recs, err := SCClient(root.Contract.Hname()).CallView(root.FuncGetContractRecords.Name, nil)
			log.Check(err)
			contracts, err := root.DecodeContractRegistry(collections.NewMapReadOnly(recs, root.StateVarContractRegistry))
			log.Check(err)
			log.Printf("#Contracts: %d\n", len(contracts))

			ownerID, err := codec.DecodeAgentID(info.MustGet(governance.VarChainOwnerID))
			log.Check(err)
			log.Printf("Owner: %s\n", ownerID.String())

			if info.MustHas(governance.VarChainOwnerIDDelegated) {
				delegated, err := codec.DecodeAgentID(info.MustGet(governance.VarChainOwnerIDDelegated))
				log.Check(err)
				log.Printf("Delegated owner: %s\n", delegated.String())
			}

			feeColor, defaultOwnerFee, defaultValidatorFee, err := governance.GetDefaultFeeInfo(info)
			log.Check(err)
			log.Printf("Default owner fee: %d %s\n", defaultOwnerFee, feeColor.String())
			log.Printf("Default validator fee: %d %s\n", defaultValidatorFee, feeColor.String())
		}
	},
}
