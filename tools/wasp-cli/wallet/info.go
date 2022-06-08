package wallet

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/tools/wasp-cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
	"github.com/spf13/cobra"
)

var addressCmd = &cobra.Command{
	Use:   "address",
	Short: "Show the wallet address",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		wallet := Load()
		log.Printf("Address index %d\n", addressIndex)
		log.Verbosef("  Private key: %s\n", wallet.KeyPair.GetPrivateKey().AsString())
		log.Verbosef("  Public key:  %s\n", wallet.KeyPair.GetPublicKey().AsString())
		if parameters.L1 == nil {
			config.L1Client() // this will fill parameters.L1 with data from the L1 node
		}
		log.Printf("  Address:     %s\n", wallet.Address().Bech32(parameters.L1.Protocol.Bech32HRP))
	},
}

var balanceCmd = &cobra.Command{
	Use:   "balance",
	Short: "Show the wallet balance",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		wallet := Load()
		address := wallet.Address()

		outs, err := config.L1Client().OutputMap(address)
		log.Check(err)

		log.Printf("Address index %d\n", addressIndex)
		log.Printf("  Address: %s\n", address.Bech32(parameters.L1.Protocol.Bech32HRP))
		log.Printf("  Balance:\n")
		if log.VerboseFlag {
			printOutputsByOutputID(outs)
		} else {
			printOutputsByTokenID(outs)
		}
	},
}

func printOutputsByTokenID(outs map[iotago.OutputID]iotago.Output) {
	balance := iscp.FungibleTokensFromOutputMap(outs)
	log.Printf("    iota: %d\n", balance.Iotas)
	for _, nt := range balance.Tokens {
		log.Printf("    %s: %s\n", nt.ID, nt.Amount)
	}
}

func printOutputsByOutputID(outs map[iotago.OutputID]iotago.Output) {
	for i, out := range outs {
		log.Printf("    output index %d:\n", i)
		tokens := iscp.FungibleTokensFromOutput(out)
		log.Printf("%s\n", tokens.String())
	}
}
