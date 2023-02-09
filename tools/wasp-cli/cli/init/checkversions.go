package init

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/iotaledger/wasp/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

func initCheckVersionsCmd(waspVersion string) *cobra.Command {
	return &cobra.Command{
		Use:   "check-versions",
		Short: "checks the versions of wasp-cli and wasp nodes match",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			// query every wasp node info endpoint and ensure the `Version` matches
			waspConfig := viper.Sub("wasp").AllSettings()
			for nodeName := range waspConfig {
				version, _, err := cliclients.WaspClient(nodeName).NodeApi.
					GetVersion(context.Background()).
					Execute()
				log.Check(err)

				if waspVersion == version.Version {
					log.Printf("Wasp-cli version matches Wasp {%s}\n", nodeName)
				} else {
					log.Printf("! -> Version mismatch with Wasp {%s}. cli version: %s, wasp version: %s\n", nodeName, waspVersion, version.Version)
				}
			}
		},
	}
}
