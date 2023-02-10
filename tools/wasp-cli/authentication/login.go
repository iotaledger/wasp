package authentication

import (
	"bufio"
	"context"
	"os"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/tools/wasp-cli/waspcmd"
)

var (
	username string
	password string
)

func initLoginCmd() *cobra.Command {
	var node string
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate against a Wasp node",
		// Args:  cobra.ArbitraryArgs,
		Run: func(cmd *cobra.Command, args []string) {
			node = waspcmd.DefaultSingleNodeFallback(node)
			if username == "" || password == "" {
				scanner := bufio.NewScanner(os.Stdin)

				log.Printf("Username: ")
				scanner.Scan()
				username = scanner.Text()

				log.Printf("Password: ")
				passwordBytes, err := term.ReadPassword(int(syscall.Stdin)) //nolint:nolintlint,unconvert // int cast is needed for windows
				if err != nil {
					panic(err)
				}

				password = string(passwordBytes)
			}

			// If credentials are still empty, exit early.
			if username == "" || password == "" {
				log.Printf("Invalid credentials")
				return
			}

			token, _, err := cliclients.WaspClient(node).AuthApi.
				Authenticate(context.Background()).
				LoginRequest(apiclient.LoginRequest{
					Username: username,
					Password: password,
				}).Execute()

			log.Check(err)

			config.SetToken(token.Jwt)

			log.Printf("\nSuccessfully authenticated\n")
		},
	}
	waspcmd.WithSingleWaspNodesFlag(cmd, &node)
	return cmd
}
