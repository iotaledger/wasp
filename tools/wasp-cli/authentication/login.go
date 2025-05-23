// Package authentication provides functionality for authenticating with Wasp nodes,
// allowing users to login, set tokens, and manage authentication information.
package authentication

import (
	"bufio"
	"context"
	"fmt"
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

func initSetTokenCmd() *cobra.Command {
	var node string
	cmd := &cobra.Command{
		Use:   "set-token",
		Short: "Manually sets a token for a given node",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			node = waspcmd.DefaultWaspNodeFallback(node)

			config.SetToken(node, args[0])

			fmt.Printf("Set token for %s", node)
		},
	}
	waspcmd.WithWaspNodeFlag(cmd, &node)
	return cmd
}

func initLoginCmd() *cobra.Command {
	var node string
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate against a Wasp node",
		// Args:  cobra.ArbitraryArgs,
		Run: func(cmd *cobra.Command, args []string) {
			node = waspcmd.DefaultWaspNodeFallback(node)
			if username == "" || password == "" {
				scanner := bufio.NewScanner(os.Stdin)

				log.Printf("Username: ")
				scanner.Scan()
				username = scanner.Text()

				log.Printf("Password: ")
				// int cast is needed for windows
				passwordBytes, err := term.ReadPassword(int(syscall.Stdin)) //nolint:unconvert
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

			ctx := context.Background()
			token, _, err := cliclients.WaspClientWithVersionCheck(ctx, node).AuthAPI.
				Authenticate(context.Background()).
				LoginRequest(apiclient.LoginRequest{
					Username: username,
					Password: password,
				}).Execute()

			log.Check(err)

			config.SetToken(node, token.Jwt)

			log.Printf("\nSuccessfully authenticated\n")
		},
	}
	waspcmd.WithWaspNodeFlag(cmd, &node)
	return cmd
}
