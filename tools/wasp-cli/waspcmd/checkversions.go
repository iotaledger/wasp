package waspcmd

import (
	"context"

	"github.com/iotaledger/wasp/components/app"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/spf13/cobra"
)

func initCheckVersionsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "check-versions",
		Short: "checks the versions of wasp-cli and wasp nodes match",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			// query every wasp node info endpoint and ensure the `Version` matches
			waspSettings := map[string]interface{}{}
			waspKey := config.Config.Cut("wasp")

			if waspKey != nil {
				waspSettings = waspKey.All()
			}
			if len(waspSettings) == 0 {
				log.Fatalf("no wasp node configured, you can add a node with `wasp-cli wasp add <name> <api url>`")
			}
			for nodeName := range waspSettings {
				ctx := context.Background()

				nodeVersion, _, err := cliclients.WaspClientWithVersionCheck(ctx, nodeName).NodeAPI.
					GetVersion(ctx).
					Execute()
				log.Check(err)
				if app.Version == "v"+nodeVersion.Version {
					log.Printf("Wasp-cli version matches Wasp {%s}\n", nodeName)
				} else {
					log.Printf("! -> Version mismatch with Wasp {%s}. cli version: %s, wasp version: %s\n", nodeName, app.Version, nodeVersion.Version)
				}
			}
		},
	}
}
