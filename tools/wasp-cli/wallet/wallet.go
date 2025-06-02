// Package wallet provides commands for managing and interacting with IOTA wallets,
// enabling users to perform various cryptocurrency operations.
package wallet

import (
	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/wallet"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

type WalletConfig struct {
	KeyPair cryptolib.Signer
}

var initOverwrite bool

func initInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a new wallet",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			wallet.InitWallet(initOverwrite)

			config.SetWalletProviderString(string(wallet.GetWalletProvider()))
			log.Check(config.WriteConfig())
		},
	}
	cmd.Flags().BoolVar(&initOverwrite, "overwrite", false, "allow overwriting existing seed")
	return cmd
}

type InitModel struct {
	Scheme string
}

var _ log.CLIOutput = &InitModel{}

func (i *InitModel) AsText() (string, error) {
	template := `Initialized wallet seed in {{ .Scheme }}`
	return log.ParseCLIOutputTemplate(i, template)
}
