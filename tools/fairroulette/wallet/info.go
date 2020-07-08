package wallet

import (
	"fmt"

	nodeapi "github.com/iotaledger/goshimmer/dapps/waspconn/packages/apilib"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/tools/fairroulette/config"
)

func dumpAddress() {
	wallet := Load()
	kp := wallet.KeyPair()
	fmt.Printf("Account index %d\n", wallet.AccountIndex())
	fmt.Printf("  Private key: %s\n", kp.PrivateKey)
	fmt.Printf("  Public key:  %s\n", kp.PublicKey)
	fmt.Printf("  Address:     %s\n", wallet.Address())
}

func dumpBalance() {
	wallet := Load()
	address := wallet.Address()

	r, err := nodeapi.GetAccountOutputs(config.GoshimmerApi(), &address)
	check(err)

	byColor, total := util.OutputBalancesByColor(r)

	fmt.Printf("Account index %d\n", wallet.AccountIndex())
	fmt.Printf("  Address: %s\n", address)
	fmt.Printf("  Balance:\n")
	for color, value := range byColor {
		fmt.Printf("    %s: %d\n", color.String(), value)
	}
	fmt.Printf("    ------\n")
	fmt.Printf("    Total: %d\n", total)
}
