package wallet

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

type WalletConfig struct {
	Seed []byte
}

func initInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize a new wallet",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			seed := cryptolib.NewSeed()
			seedString := iotago.EncodeHex(seed[:])
			viper.Set("wallet.seed", seedString)
			log.Check(viper.WriteConfig())

			model := &InitModel{
				ConfigPath: config.ConfigPath,
			}

			if log.VerboseFlag {
				verboseOutputs := make(map[string]string)
				verboseOutputs["Seed"] = seedString
				model.VerboseOutputs = verboseOutputs
			}

			log.PrintCLIOutput(model)
		},
	}
}

type InitModel struct {
	ConfigPath     string
	Message        string
	VerboseOutputs map[string]string
}

var _ log.CLIOutput = &InitModel{}

func (i *InitModel) AsText() (string, error) {
	template := `Initialized wallet seed in {{ .ConfigPath }}
IMPORTANT: wasp-cli is alpha phase. The seed is currently being stored in a plain text file which is NOT secure. Do not use this seed to store funds in the mainnet

  {{ range $i, $out := .VerboseOutputs }}
    {{ $i }}: {{ $out}}
  {{ end }}`
	return log.ParseCLIOutputTemplate(i, template)
}
