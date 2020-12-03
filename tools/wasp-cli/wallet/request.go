package wallet

import (
	"github.com/iotaledger/wasp/tools/wasp-cli/config"
)

func requestFundsCmd(args []string) {
	address := Load().Address()
	// automatically waits for confirmation:
	check(config.GoshimmerClient().RequestFunds(&address))
}
