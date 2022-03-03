package authentication

import (
	"encoding/json"

	"github.com/iotaledger/wasp/tools/wasp-cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info",
	Short: "Receive information about the authentication methods",
	Run: func(cmd *cobra.Command, args []string) {
		client := config.WaspClient()
		authInfo, err := client.AuthInfo()
		if err != nil {
			panic(err)
		}

		authInfoJSON, err := json.MarshalIndent(authInfo, "", "\t")
		if err != nil {
			panic(err)
		}

		log.Printf(string(authInfoJSON))
	},
}
