package authentication

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/v2/tools/wasp-cli/cli/config"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/log"
)

func initImportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import",
		Short: "Imports all JWT tokens from the config into the OS Keychain",
		Run: func(cmd *cobra.Command, args []string) {
			tokens := config.GetAuthTokenForImport()

			fmt.Println("Importing JWT tokens from the config into the OS Keychain.")

			kc := config.GetKeyChain()
			for k, v := range tokens {
				if v == "" {
					fmt.Printf("Could not import JWT token for node %q\n", k)
				} else {
					err := kc.SetJWTAuthToken(k, v)
					log.Check(err)

					fmt.Printf("Imported JWT token for node %q\n", k)
				}
			}
		},
	}
	return cmd
}
