package wallet

import (
	"github.com/iotaledger/wasp/tools/wasp-cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/spf13/cobra"
)

var requestFundsCmd = &cobra.Command{
	Use:   "request-funds",
	Short: "Request funds from the faucet",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		address := Load().Address()
		log.Check(config.L1Client().RequestFunds(address))
		log.Printf("Request funds for address %s: success\n", address.Bech32(config.L1NetworkPrefix()))
	},
}
