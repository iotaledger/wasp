package wallet

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/util"
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

	outs, err := config.GoshimmerClient().GetConfirmedOutputs(address)
	log.Check(err)

	log.Printf("Address index %d\n", addressIndex)
	log.Printf("  Address: %s\n", address)
	log.Printf("  Balance:\n")
	var total uint64
	if log.VerboseFlag {
		total = byOutputId(outs)
	} else {
		total = byColor(outs)
	}
	log.Printf("    ------\n")
	log.Printf("    Total: %d\n", total)
}

func byColor(outs []ledgerstate.Output) uint64 {
	byColor, total := util.OutputBalancesByColor(outs)
	for color, value := range byColor {
		log.Printf("    %s: %d\n", color.String(), value)
	}
	return total
}

func byOutputId(outs []ledgerstate.Output) uint64 {
	var total uint64
	for _, out := range outs {
		log.Printf("    output ID %s:\n", out.ID())
		out.Balances().ForEach(func(color ledgerstate.Color, balance uint64) bool {
			log.Printf("      %s: %d\n", color, balance)
			total += balance
			return true
		})
	}
	return total
}
