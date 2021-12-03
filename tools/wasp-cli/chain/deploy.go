package chain

import (
	"os"

	"github.com/iotaledger/wasp/packages/apilib"
	"github.com/iotaledger/wasp/packages/iscp"
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
				if committee != nil {
					peers = committee
				} else {
					peers = []int{0, 1, 2, 3}
				}
			}

			if committee == nil {
				committee = peers
			}

			chainid, _, err := apilib.DeployChainWithDKG(apilib.CreateChainParams{
				Layer1Client:          config.GoshimmerClient(),
				AllAPIHosts:           config.CommitteeAPI(peers),
				AllPeeringHosts:       config.CommitteePeering(peers),
				CommitteeAPIHosts:     config.CommitteeAPI(committee),
				CommitteePeeringHosts: config.CommitteePeering(committee),
				N:                     uint16(len(committee)),
				T:                     uint16(quorum),
				OriginatorPrivateKey:  wallet.Load().KeyPair(),
				Description:           description,
				Textout:               os.Stdout,
			})
			log.Check(err)

			AddChainAlias(alias, chainid.Bech32(iscp.Bech32Prefix))
		},
	}

	cmd.Flags().IntSliceVarP(&peers, "peers", "", nil, "indices of peer nodes (default: 0,1,2,3)")
	cmd.Flags().IntSliceVarP(&committee, "committee", "", nil, "subset of peers acting as committee nodes  (default: same as peers)")
	cmd.Flags().IntVarP(&quorum, "quorum", "", 3, "quorum")
	cmd.Flags().StringVarP(&description, "description", "", "", "description")
	return cmd
}
