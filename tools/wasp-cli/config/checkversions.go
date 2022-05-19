package config

import (
	"github.com/iotaledger/wasp/client"
	"github.com/iotaledger/wasp/packages/wasp"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/spf13/cobra"
)

var checkVersionsCmd = &cobra.Command{
	Use:   "check-versions",
	Short: "checks the versions of wasp-cli and wasp nodes match",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		// query every wasp node info endpoint and ensure the `VersionHash` matches
		for i := 0; i < totalNumberOfWaspNodes(); i++ {
			client := client.NewWaspClient(committeeHost(HostKindAPI, i))
			waspServerInfo, err := client.Info()
			log.Check(err)
			if wasp.VersionHash == waspServerInfo.VersionHash {
				log.Printf("Wasp-cli version matches Wasp #%d\n", i)
			} else {
				log.Printf("! -> Version mismatch with Wasp #%d. cli hash: %s, wasp hash: %s\n", i, wasp.VersionHash, waspServerInfo.VersionHash)
			}
		}
	},
}
