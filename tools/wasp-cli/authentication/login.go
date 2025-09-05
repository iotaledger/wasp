// Package authentication provides functionality for authenticating with Wasp nodes,
// allowing users to login, set tokens, and manage authentication information.
package authentication

import (
	"bufio"
	"context"
	"os"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/iotaledger/wasp/v2/clients/apiclient"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/cli/config"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/format"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/waspcmd"
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
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			node, err = waspcmd.DefaultWaspNodeFallback(node)
			if err != nil {
				return err
			}

			config.SetToken(node, args[0])

			authOutput := format.NewAuthSuccess(node, "manual")
			authOutput.Data.Message = "Token set successfully"
			return format.PrintOutput(authOutput)
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
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			node, err = waspcmd.DefaultWaspNodeFallback(node)
			if err != nil {
				return err
			}
			if username == "" || password == "" {
				scanner := bufio.NewScanner(os.Stdin)

				log.Printf("Username: ")
				scanner.Scan()
				username = scanner.Text()

				log.Printf("Password: ")
				// int cast is needed for windows
				var passwordBytes []byte
				passwordBytes, err = term.ReadPassword(int(syscall.Stdin)) //nolint:unconvert
				if err != nil {
					return err
				}

				password = string(passwordBytes)
			}

			// If credentials are still empty, exit early.
			if username == "" || password == "" {
				authOutput := format.NewAuthError(node, username, "Invalid credentials provided")
				return format.PrintOutput(authOutput)
			}

			ctx := context.Background()
			token, _, err := cliclients.WaspClientWithVersionCheck(ctx, node).AuthAPI.
				Authenticate(context.Background()).
				LoginRequest(apiclient.LoginRequest{
					Username: username,
					Password: password,
				}).Execute()
			if err != nil {
				authOutput := format.NewAuthError(node, username, err.Error())
				return format.PrintOutput(authOutput)
			}

			config.SetToken(node, token.Jwt)

			authOutput := format.NewAuthSuccess(node, username)
			return format.PrintOutput(authOutput)
		},
	}
	waspcmd.WithWaspNodeFlag(cmd, &node)
	return cmd
}
