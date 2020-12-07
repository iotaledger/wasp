package wallet

import (
	"github.com/iotaledger/wasp/tools/wasp-cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

func requestFundsCmd(args []string) {
	address := Load().Address()
	// automatically waits for confirmation:
	log.Check(config.GoshimmerClient().RequestFunds(&address))
	log.Printf("Request funds for address %s: success\n", address)
}
