package chain

import (
	"context"
	"net/http"

	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/tools/wasp-cli/waspcmd"
)

func initActivateCmd() *cobra.Command {
	var node string
	var chain string
	cmd := &cobra.Command{
		Use:   "activate",
		Short: "Activates the chain on selected nodes",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			node = waspcmd.DefaultWaspNodeFallback(node)
			chain = defaultChainFallback(chain)
			chainID := config.GetChain(chain)
			activateChain(node, chainID)
		},
	}

	waspcmd.WithWaspNodeFlag(cmd, &node)

	withChainFlag(cmd, &chain)
	return cmd
}

func activateChain(node string, chainID isc.ChainID) {
	client := cliclients.WaspClient(node)
	r, httpStatus, err := client.ChainsApi.GetChainInfo(context.Background(), chainID.String()).Execute() //nolint:bodyclose // false positive

	if err != nil && httpStatus.StatusCode != http.StatusNotFound {
		log.Check(err)
	}

	if r != nil && r.IsActive {
		return
	}

	if r == nil {
		_, err2 := client.ChainsApi.SetChainRecord(context.Background(), chainID.String()).ChainRecord(apiclient.ChainRecord{
			IsActive:    true,
			AccessNodes: []string{},
		}).Execute() //nolint:bodyclose // false positive

		log.Check(err2)
	} else {
		_, err = client.ChainsApi.ActivateChain(context.Background(), chainID.String()).Execute() //nolint:bodyclose // false positive

		log.Check(err)
	}

	log.Printf("Chain activated")
}

func initDeactivateCmd() *cobra.Command {
	var node string
	var chain string

	cmd := &cobra.Command{
		Use:   "deactivate",
		Short: "Deactivates the chain on selected nodes",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			chain = defaultChainFallback(chain)

			chainID := config.GetChain(chain)
			node = waspcmd.DefaultWaspNodeFallback(node)
			client := cliclients.WaspClient(node)
			_, err := client.ChainsApi.DeactivateChain(context.Background(), chainID.String()).Execute() //nolint:bodyclose // false positive
			log.Check(err)
		},
	}
	waspcmd.WithWaspNodeFlag(cmd, &node)
	withChainFlag(cmd, &chain)
	return cmd
}
