package authentication

import (
	"encoding/json"

	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/tools/wasp-cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

func initInfoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: "Receive information about the authentication methods",
		Run: func(cmd *cobra.Command, args []string) {
			client := config.WaspClient(config.MustWaspAPI())
			authInfo, err := client.AuthInfo()
			if err != nil {
				panic(err)
			}

			authInfoJSON, err := json.MarshalIndent(authInfo, "", "  ")
			if err != nil {
				panic(err)
			}

			log.Printf(string(authInfoJSON))
		},
	}
}
