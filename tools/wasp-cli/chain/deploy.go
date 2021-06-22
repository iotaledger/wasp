package chain

import (
	"os"

	"github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/tools/wasp-cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/tools/wasp-cli/wallet"
	"github.com/spf13/cobra"
)

func deployCmd() *cobra.Command {
	var (
		peers       []int
		committee   []int
		quorum      int
		description string
	)

	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy a new chain",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			alias := GetChainAlias()

			if peers == nil {
				peers = committee
			}

			chainid, _, err := apilib.DeployChainWithDKG(apilib.CreateChainParams{
				Node:                  config.GoshimmerClient(),
				AllApiHosts:           config.CommitteeApi(peers),
				AllPeeringHosts:       config.CommitteePeering(peers),
				CommitteeApiHosts:     config.CommitteeApi(committee),
				CommitteePeeringHosts: config.CommitteePeering(committee),
				N:                     uint16(len(committee)),
				T:                     uint16(quorum),
				OriginatorKeyPair:     wallet.Load().KeyPair(),
				Description:           description,
				Textout:               os.Stdout,
			})
			log.Check(err)

			AddChainAlias(alias, chainid.Base58())
		},
	}

	cmd.Flags().IntSliceVarP(&committee, "committee", "", []int{0, 1, 2, 3}, "indices of committee nodes")
	cmd.Flags().IntSliceVarP(&committee, "peers", "", nil, "indices of peer nodes (default: same as committee)")
	cmd.Flags().IntVarP(&quorum, "quorum", "", 3, "quorum")
	cmd.Flags().StringVarP(&description, "description", "", "", "description")
	return cmd
}
