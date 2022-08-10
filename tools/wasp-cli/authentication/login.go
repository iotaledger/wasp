package authentication

import (
	"bufio"
	"os"
	"syscall"

	"github.com/iotaledger/wasp/tools/wasp-cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var (
	username string
	password string
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate against a Wasp node",
	// Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if username == "" || password == "" {
			scanner := bufio.NewScanner(os.Stdin)

			log.Printf("Username: ")
			scanner.Scan()
			username = scanner.Text()

			log.Printf("Password: ")
			passwordBytes, err := term.ReadPassword(int(syscall.Stdin)) //nolint:unconvert // int cast is needed for windows
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

		client := config.WaspClient()
		token, err := client.Login(username, password)
		if err != nil {
			panic(err)
		}

		config.SetToken(token)

		log.Printf("\nSuccessfully authorized\n")
	},
}
