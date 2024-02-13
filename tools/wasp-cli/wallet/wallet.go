package wallet

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/wallet"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

type WalletConfig struct {
	KeyPair cryptolib.VariantKeyPair
}

func initInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize a new wallet",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			wallet.InitWallet()

			config.SetWalletProviderString(string(wallet.GetWalletProvider()))
			log.Check(viper.WriteConfig())
		},
	}
}

type InitModel struct {
	Scheme string
}

var _ log.CLIOutput = &InitModel{}

func (i *InitModel) AsText() (string, error) {
	template := `Initialized wallet seed in {{ .Scheme }}`
	return log.ParseCLIOutputTemplate(i, template)
}
