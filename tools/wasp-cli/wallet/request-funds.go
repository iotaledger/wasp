package wallet

import (
	"github.com/spf13/cobra"

	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/tools/wasp-cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

var requestFundsCmd = &cobra.Command{
	Use:   "request-funds",
	Short: "Request funds from the faucet",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		address := Load().Address()
		log.Check(config.L1Client().RequestFunds(address))
		log.Printf("Request funds for address %s: success\n", address.Bech32(parameters.L1().Protocol.Bech32HRP))
	},
}
