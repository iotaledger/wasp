package chain

import (
	"context"
	"net/http"

	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

func initActivateCmd() *cobra.Command {
	var nodes []int
	cmd := &cobra.Command{
		Use:   "activate",
		Short: "Activates the chain on selected nodes",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			chainID := config.GetCurrentChainID()
			if nodes == nil {
				nodes = GetAllWaspNodes()
			}
			for _, nodeIdx := range nodes {
				client := cliclients.WaspClientForIndex(nodeIdx)

				r, httpStatus, err := client.ChainsApi.GetChainInfo(context.Background(), chainID.String()).Execute()

				if err != nil && httpStatus.StatusCode != http.StatusNotFound {
					log.Check(err)
				}

				if r != nil && r.IsActive {
					continue
				}

				if r == nil {
					_, err := client.ChainsApi.SetChainRecord(context.Background(), chainID.String()).ChainRecord(apiclient.ChainRecord{
						IsActive:    true,
						AccessNodes: []string{},
					}).Execute()

					log.Check(err)
				} else {
					_, err = client.ChainsApi.ActivateChain(context.Background(), chainID.String()).Execute()

					log.Check(err)
				}
			}

			log.Printf("Chain activated")
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
			chainID := config.GetCurrentChainID()
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
