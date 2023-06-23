package wallet

import (
	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/packages/cryptolib"
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
			scheme := wallet.GetWalletScheme()
			log.Printf(scheme)
			if scheme != wallet.SchemeKeyChain {
				log.Fatal("Nothing to do here")
				return
			}

			keyChain := wallet.NewKeyChain()
			err := keyChain.InitializeKeyPair()
			log.Check(err)

			model := &InitModel{
				Scheme: scheme,
			}

			log.PrintCLIOutput(model)
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
