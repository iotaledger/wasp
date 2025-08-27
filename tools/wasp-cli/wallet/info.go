package wallet

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/cli/cliclients"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/cli/wallet"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/format"
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
			addressOutput := format.NewWalletAddressSuccess(myWallet.AddressIndex(), address.String())
			err := format.PrintOutput(addressOutput)
			if err != nil {
				log.Printf("Error formatting output: %v", err)
			}
		},
	}
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
			if err != nil {
				balanceOutput := format.NewWalletBalanceError(myWallet.AddressIndex(), address.String())
				formatErr := format.PrintOutput(balanceOutput)
				if formatErr != nil {
					log.Printf("Error formatting output: %v", formatErr)
				}
				return
			}

			balanceOutput := format.NewWalletBalanceSuccess(myWallet.AddressIndex(), address.String(), balance)
			err = format.PrintOutput(balanceOutput)
			if err != nil {
				log.Printf("Error formatting output: %v", err)
			}
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
