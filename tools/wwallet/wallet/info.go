package wallet

import (
	"fmt"

	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/balance"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/tools/wwallet/config"
)

func dumpAddress() {
	wallet := Load()
	kp := wallet.KeyPair()
	fmt.Printf("Address index %d\n", addressIndex)
	fmt.Printf("  Private key: %s\n", kp.PrivateKey)
	fmt.Printf("  Public key:  %s\n", kp.PublicKey)
	fmt.Printf("  Address:     %s\n", wallet.Address())
}

func dumpBalance() {
	wallet := Load()
	address := wallet.Address()

	outs, err := config.GoshimmerClient().GetAccountOutputs(&address)
	check(err)

	fmt.Printf("Address index %d\n", addressIndex)
	fmt.Printf("  Address: %s\n", address)
	fmt.Printf("  Balance:\n")
	var total int64
	if config.Verbose {
		total = byOutputId(outs)
	} else {
		total = byColor(outs)
	}
	fmt.Printf("    ------\n")
	fmt.Printf("    Total: %d\n", total)
}

func byColor(outs map[valuetransaction.OutputID][]*balance.Balance) int64 {
	byColor, total := util.OutputBalancesByColor(outs)
	for color, value := range byColor {
		fmt.Printf("    %s: %d\n", color.String(), value)
	}
	return total
}

func byOutputId(outs map[valuetransaction.OutputID][]*balance.Balance) int64 {
	var total int64
	for outputID, bals := range outs {
		fmt.Printf("    output ID %s:\n", outputID)
		for _, bal := range bals {
			fmt.Printf("      %s: %d\n", bal.Color.String(), bal.Value)
		}
	}
	return total
}
