package wallet

import (
	"github.com/iotaledger/wasp/packages/txutil"
	"github.com/iotaledger/wasp/tools/wasp-cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

func addressCmd(args []string) {
	wallet := Load()
	kp := wallet.KeyPair()
	log.Printf("Address index %d\n", addressIndex)
	log.Verbose("  Private key: %s\n", kp.PrivateKey)
	log.Verbose("  Public key:  %s\n", kp.PublicKey)
	log.Printf("  Address:     %s\n", wallet.Address())
}

func balanceCmd(args []string) {
	wallet := Load()
	address := wallet.Address()

	outs, err := config.GoshimmerClient().GetConfirmedOutputs(&address)
	log.Check(err)

	log.Printf("Address index %d\n", addressIndex)
	log.Printf("  Address: %s\n", address)
	log.Printf("  Balance:\n")
	var total int64
	if log.VerboseFlag {
		total = byOutputId(outs)
	} else {
		total = byColor(outs)
	}
	log.Printf("    ------\n")
	log.Printf("    Total: %d\n", total)
}

func byColor(outs map[valuetransaction.OutputID][]*balance.Balance) int64 {
	byColor, total := txutil.OutputBalancesByColor(outs)
	for color, value := range byColor {
		log.Printf("    %s: %d\n", color.String(), value)
	}
	return total
}

func byOutputId(outs map[valuetransaction.OutputID][]*balance.Balance) int64 {
	var total int64
	for outputID, bals := range outs {
		log.Printf("    output ID %s:\n", outputID)
		for _, bal := range bals {
			log.Printf("      %s: %d\n", bal.Color.String(), bal.Value)
			total += bal.Value
		}
	}
	return total
}
