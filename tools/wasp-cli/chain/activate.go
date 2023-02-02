package chain

import (
	"context"
	"regexp"

	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

var HTTP404ErrRegexp = regexp.MustCompile(`"Code":404`)

func initActivateCmd() *cobra.Command {
	var nodes []int
	cmd := &cobra.Command{
		Use:   "activate",
		Short: "Activates the chain on selected nodes",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			chainID := GetCurrentChainID()
			if nodes == nil {
				nodes = GetAllWaspNodes()
			}
			for _, nodeIdx := range nodes {
				client := cliclients.WaspClientForIndex(nodeIdx)

				r, _, err := client.ChainsApi.GetChainInfo(context.Background(), chainID.String()).Execute()

				if err != nil && !HTTP404ErrRegexp.MatchString(err.Error()) {
					log.Check(err)
				}

				if r != nil && r.IsActive {
					continue
				}

				_, err = client.ChainsApi.ActivateChain(context.Background(), chainID.String()).Execute()
				log.Check(err)
			}
		},
	}

	cmd.Flags().IntSliceVarP(&nodes, "nodes", "", nil, "nodes to activate the chain on (ex: 0,1,2,3) (default: all nodes)")

	return cmd
}

func initDeactivateCmd() *cobra.Command {
	var nodes []int
	cmd := &cobra.Command{
		Use:   "deactivate",
		Short: "Deactivates the chain on selected nodes",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			chainID := GetCurrentChainID()
			if nodes == nil {
				nodes = GetAllWaspNodes()
			}
			for _, nodeIdx := range nodes {
				client := cliclients.WaspClientForIndex(nodeIdx)

				_, err := client.ChainsApi.DeactivateChain(context.Background(), chainID.String()).Execute()
				log.Check(err)
			}
		},
	}
	cmd.Flags().IntSliceVarP(&nodes, "nodes", "", nil, "nodes to deactivate the chain on (ex: 0,1,2,3) (default: all nodes)")
	return cmd
}
