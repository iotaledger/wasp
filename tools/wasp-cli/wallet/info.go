package wallet

import (
	"context"
	"fmt"

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
			err := format.FormatWalletAddress(myWallet.AddressIndex(), address.String())
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
				formatErr := format.FormatError("wallet_balance", fmt.Sprintf("Address: %s (index %d), Error: %s", address.String(), myWallet.AddressIndex(), err.Error()))
				if formatErr != nil {
					log.Printf("Error formatting output: %v", formatErr)
				}
				return
			}

			err = format.FormatWalletBalance(myWallet.AddressIndex(), address.String(), balance)
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
