package wallet

import (
	"github.com/iotaledger/wasp/tools/wasp-client/config"
)

func requestFunds() {
	address := Load().Address()
	check(config.GoshimmerClient().RequestFunds(&address))
}
