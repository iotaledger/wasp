package wallet

import (
	"github.com/iotaledger/wasp/tools/wasp-client/config"
)

func requestFunds() {
	address := Load().Address()
	// automatically waits for confirmation:
	check(config.GoshimmerClient().RequestFunds(&address))
}
