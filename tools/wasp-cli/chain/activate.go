package chain

import (
	"context"
	"net/http"

	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/v2/clients/apiclient"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/cli/config"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/waspcmd"
)

func initActivateCmd() *cobra.Command {
	var node string
	var chainName string
	cmd := &cobra.Command{
		Use:   "activate",
		Short: "Activates the chain on selected nodes",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			node = waspcmd.DefaultWaspNodeFallback(node)
			chainName = defaultChainFallback(chainName)
			chainID := config.GetChain(chainName)
			ctx := context.Background()
			activateChain(ctx, node, chainName, chainID)
		},
	}

	waspcmd.WithWaspNodeFlag(cmd, &node)

	withChainFlag(cmd, &chainName)
	return cmd
}

func activateChain(ctx context.Context, node string, chainName string, chainID isc.ChainID) {
	client := cliclients.WaspClientWithVersionCheck(ctx, node)
	r, httpStatus, err := client.ChainsAPI.GetChainInfo(ctx).Execute() //nolint:bodyclose // false positive

	if err != nil && httpStatus.StatusCode != http.StatusNotFound {
		log.Check(err)
	}

	if r != nil && r.IsActive {
		return
	}

	if r == nil {
		_, err2 := client.ChainsAPI.SetChainRecord(ctx, chainID.String()).ChainRecord(apiclient.ChainRecord{
			IsActive:    true,
			AccessNodes: []string{},
		}).Execute() //nolint:bodyclose // false positive

		log.Check(err2)
	} else {
		_, err = client.ChainsAPI.ActivateChain(ctx, chainID.String()).Execute() //nolint:bodyclose // false positive
		log.Check(err)
	}

	log.Printf("Chain: %v (%v)\nActivated\n", chainID, chainName)
}

func initDeactivateCmd() *cobra.Command {
	var node string
	var chainName string

	cmd := &cobra.Command{
		Use:   "deactivate",
		Short: "Deactivates the chain on selected nodes",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			chainName = defaultChainFallback(chainName)

			node = waspcmd.DefaultWaspNodeFallback(node)
			ctx := context.Background()
			client := cliclients.WaspClientWithVersionCheck(ctx, node)
			_, err := client.ChainsAPI.DeactivateChain(ctx).Execute() //nolint:bodyclose // false positive
			log.Check(err)
			log.Printf("Chain: %v \nDeactivated.\n", chainName)
		},
	}
	waspcmd.WithWaspNodeFlag(cmd, &node)
	withChainFlag(cmd, &chainName)
	return cmd
}
