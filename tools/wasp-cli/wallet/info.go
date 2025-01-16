package wallet

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/wallet"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

func initAddressCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "address",
		Short: "Show the wallet address",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			myWallet := wallet.Load()
			address := myWallet.Address()

			model := &AddressModel{Address: address.String(), Index: int(myWallet.AddressIndex())}

			if log.VerboseFlag {
				verboseOutput := make(map[string]string)
				// verboseOutput["Private key"] = myWallet.KeyPair.GetPrivateKey().String()
				// verboseOutput["Public key"] = myWallet.GetPublicKey().String() // TODO: is it needed?
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
			myWallet := wallet.Load()
			address := myWallet.Address()

			balance, err := cliclients.L1Client().GetAllBalances(context.Background(), address.AsIotaAddress())
			log.Check(err)

			model := &BalanceModel{
				Address:      address.String(),
				AddressIndex: myWallet.AddressIndex(),
				Tokens:       balance,
			}

			log.PrintCLIOutput(model)
		},
	}
}

var _ log.CLIOutput = &BalanceModel{}

type BalanceModel struct {
	AddressIndex uint32                 `json:"AddressIndex"`
	Address      string                 `json:"Address"`
	Tokens       []*iotajsonrpc.Balance `json:"BaseTokens"`
}

func (b *BalanceModel) AsText() (string, error) {
	balanceTemplate := `Address index: {{.AddressIndex}}
Address: {{.Address}}

Native Assets:

 {{range $i, $out := .Tokens}}
 - {{$out.CoinType}}: {{$out.TotalBalance}}
{{end}}`

	return log.ParseCLIOutputTemplate(b, balanceTemplate)
}
