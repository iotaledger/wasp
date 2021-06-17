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

			chainid, _, err := apilib.DeployChainWithDKG(apilib.CreateChainParams{
				Node:                  config.GoshimmerClient(),
				AllApiHosts:           config.CommitteeApi(committee),
				AllPeeringHosts:       config.CommitteePeering(committee),
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
	cmd.Flags().IntSliceVarP(&committee, "committee", "", []int{0, 1, 2, 3}, "committee indices")
	cmd.Flags().IntVarP(&quorum, "quorum", "", 3, "quorum")
	cmd.Flags().StringVarP(&description, "description", "", "", "description")
	return cmd
}
