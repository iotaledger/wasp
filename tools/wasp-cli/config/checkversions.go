package config

import (
	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

func initCheckVersionsCmd(waspVersion string) *cobra.Command {
	return &cobra.Command{
		Use:   "check-versions",
		Short: "checks the versions of wasp-cli and wasp nodes match",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			// query every wasp node info endpoint and ensure the `Version` matches
			for i := 0; i < totalNumberOfWaspNodes(); i++ {
				client := client.NewWaspClient(committeeHost(HostKindAPI, i))
				waspServerInfo, err := client.Info()
				log.Check(err)
				if waspVersion == waspServerInfo.Version {
					log.Printf("Wasp-cli version matches Wasp #%d\n", i)
				} else {
					log.Printf("! -> Version mismatch with Wasp #%d. cli version: %s, wasp version: %s\n", i, waspVersion, waspServerInfo.Version)
				}
			}
		},
	}
}
