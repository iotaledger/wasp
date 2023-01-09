package chain

import (
	"regexp"

	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/tools/wasp-cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

var HTTP404ErrRegexp = regexp.MustCompile(`"Code":404`)

func activateCmd() *cobra.Command {
	var nodes []int
	cmd := &cobra.Command{
		Use:   "activate",
		Short: "Activates the chain on selected nodes",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			chainID := GetCurrentChainID()
			if nodes == nil {
				nodes = getAllWaspNodes()
			}
			for _, nodeIdx := range nodes {
				client := Client(config.WaspAPI(nodeIdx))
				r, err := client.WaspClient.GetChainInfo(chainID)

				if err != nil && !HTTP404ErrRegexp.MatchString(err.Error()) {
					log.Check(err)
				}
				if r != nil && r.Active {
					continue
				}
				if r == nil {
					log.Check(
						// TODO: Activate cain on PutChain. Then the `ActivateChain` can be put to the 'else` block.`
						client.WaspClient.PutChainRecord(registry.NewChainRecord(chainID, true, []*cryptolib.PublicKey{})),
					)
				}
				log.Check(client.WaspClient.ActivateChain(chainID))
			}
		},
	}

	cmd.Flags().IntSliceVarP(&nodes, "nodes", "", nil, "nodes to activate the chain on (ex: 0,1,2,3) (default: all nodes)")

	return cmd
}

func deactivateCmd() *cobra.Command {
	var nodes []int
	cmd := &cobra.Command{
		Use:   "deactivate",
		Short: "Deactivates the chain on selected nodes",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			chainID := GetCurrentChainID()
			if nodes == nil {
				nodes = getAllWaspNodes()
			}
			for _, nodeIdx := range nodes {
				log.Check(Client(config.WaspAPI(nodeIdx)).WaspClient.DeactivateChain(chainID))
			}
		},
	}
	cmd.Flags().IntSliceVarP(&nodes, "nodes", "", nil, "nodes to deactivate the chain on (ex: 0,1,2,3) (default: all nodes)")
	return cmd
}
