package wallet

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/v2/clients/iotagraphql"
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
		RunE: func(cmd *cobra.Command, args []string) error {
			myWallet := wallet.Load()
			address := myWallet.Address()
			return format.FormatWalletAddress(myWallet.AddressIndex(), address.String())
		},
	}
}

func initBalanceCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "balance",
		Short: "Show the wallet balance",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			myWallet := wallet.Load()
			address := myWallet.Address()
			balances, err := cliclients.L1Client().GetAllBalances(context.Background(), *address.AsIotaAddress())
			if err != nil {
				// Return the error so it can be formatted by the top-level handler
				return fmt.Errorf("fetching balance for address %s (index %d): %w", address.String(), myWallet.AddressIndex(), err)
			}

			return format.FormatWalletBalance(myWallet.AddressIndex(), address.String(), balances)
		},
	}
}

var _ log.CLIOutput = &BalanceModel{}

type BalanceModel struct {
	AddressIndex uint32
	Address      string
	Balance      []*iotagraphql.Balance
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
