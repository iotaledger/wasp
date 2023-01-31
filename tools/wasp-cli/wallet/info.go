package wallet

import (
	"github.com/spf13/cobra"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/tools/wasp-cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

func initAddressCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "address",
		Short: "Show the wallet address",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			wallet := Load()

			address := wallet.Address()

			model := &AddressModel{Address: address.Bech32(parameters.L1().Protocol.Bech32HRP)}

			if log.VerboseFlag {
				verboseOutput := make(map[string]string)
				verboseOutput["Private key"] = wallet.KeyPair.GetPrivateKey().String()
				verboseOutput["Public key"] = wallet.KeyPair.GetPublicKey().String()
				model.VerboseOutput = verboseOutput
			}
			log.PrintCLIOutput(model)
		},
	}
}

type AddressModel struct {
	Index         int
	Address       string
	VerboseOutput map[string]string
}

var _ log.CLIOutput = &AddressModel{}

func (a *AddressModel) AsText() (string, error) {
	addressTemplate := `Address index: {{ .Index }}
  Address: {{ .Address }}

  {{ range $i, $out := .VerboseOutput }}
    {{ $i }}: {{ $out }}
  {{ end }}
  `
	return log.ParseCLIOutputTemplate(a, addressTemplate)
}

func initBalanceCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "balance",
		Short: "Show the wallet balance",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			wallet := Load()
			address := wallet.Address()

			outs, err := config.L1Client().OutputMap(address)
			log.Check(err)

			balance := isc.AssetsFromOutputMap(outs)

			model := &BalanceModel{
				Address:      address.Bech32(parameters.L1().Protocol.Bech32HRP),
				AddressIndex: addressIndex,
				NativeTokens: balance.NativeTokens,
				BaseTokens:   balance.BaseTokens,
				OutputMap:    outs,
			}
			if log.VerboseFlag {
				model.VerboseOutputs = map[uint16]string{}

				for i, out := range outs {
					tokens := isc.AssetsFromOutput(out)
					model.VerboseOutputs[i.Index()] = tokens.String()
				}
			}

			log.PrintCLIOutput(model)
		},
	}
}

var _ log.CLIOutput = &BalanceModel{}

type BalanceModel struct {
	AddressIndex int                 `json:"AddressIndex"`
	Address      string              `json:"Address"`
	BaseTokens   uint64              `json:"BaseTokens"`
	NativeTokens iotago.NativeTokens `json:"NativeTokens"`

	OutputMap      iotago.OutputSet `json:"-"`
	VerboseOutputs map[uint16]string
}

func (b *BalanceModel) AsText() (string, error) {
	balanceTemplate := `Address index: {{.AddressIndex}}
Address: {{.Address}}

Native Assets:

 - base: {{.BaseTokens}}{{range $i, $out := .NativeTokens}}
 - {{$out.ID}}: {{$out.Amount}}
{{end}}`

	return log.ParseCLIOutputTemplate(b, balanceTemplate)
}
