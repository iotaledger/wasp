package wallet

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/cli/wallet"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/log"
)

func initAddressCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "address",
		Short: "Show the wallet address",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			myWallet := wallet.Load()
			address := myWallet.Address()
			log.PrintCLIOutput(&AddressModel{
				Address: address.String(),
				Index:   int(myWallet.AddressIndex()),
			})
		},
	}
}

type AddressModel struct {
	Index   int
	Address string
}

var _ log.CLIOutput = &AddressModel{}

func (a *AddressModel) AsText() (string, error) {
	addressTemplate := `Address index: {{ .Index }}
Address: {{ .Address }}
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

			log.PrintCLIOutput(&BalanceModel{
				Address:      address.String(),
				AddressIndex: myWallet.AddressIndex(),
				Balance:      balance,
			})
		},
	}
}

var _ log.CLIOutput = &BalanceModel{}

type BalanceModel struct {
	AddressIndex uint32
	Address      string
	Balance      []*iotajsonrpc.Balance
}

func (b *BalanceModel) AsText() (string, error) {
	balanceTemplate := `Address index: {{.AddressIndex}}
Address: {{.Address}}

Native Assets:

 {{range $i, $out := .Balance}}
 - {{$out.CoinType}}: {{$out.TotalBalance}}
{{end}}`

	return log.ParseCLIOutputTemplate(b, balanceTemplate)
}
