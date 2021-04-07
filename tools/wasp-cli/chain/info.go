package chain

import (
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/collections"
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

		log.Printf("Chain ID: %s\n", chain.ChainID)
		log.Printf("Committee nodes: %+v\n", committee.Nodes)
		log.Printf("Active: %v\n", chain.Active)

		if chain.Active {
			info, err := SCClient(root.Interface.Hname()).CallView(root.FuncGetChainInfo)
			log.Check(err)

			description, _, err := codec.DecodeString(info.MustGet(root.VarDescription))
			log.Check(err)
			log.Printf("Description: %s\n", description)

			contracts, err := root.DecodeContractRegistry(collections.NewMapReadOnly(info, root.VarContractRegistry))
			log.Check(err)
			log.Printf("#Contracts: %d\n", len(contracts))

			ownerID, _, err := codec.DecodeAgentID(info.MustGet(root.VarChainOwnerID))
			log.Check(err)
			log.Printf("Owner: %s\n", ownerID.String())

			delegated, ok, err := codec.DecodeAgentID(info.MustGet(root.VarChainOwnerIDDelegated))
			log.Check(err)
			if ok {
				log.Printf("Delegated owner: %s\n", delegated.String())
			}

			feeColor, defaultOwnerFee, defaultValidatorFee, err := root.GetDefaultFeeInfo(info)
			log.Check(err)
			log.Printf("Default owner fee: %d %s\n", defaultOwnerFee, feeColor)
			log.Printf("Default validator fee: %d %s\n", defaultValidatorFee, feeColor)
		}
	},
}
