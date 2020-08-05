package wallet

import (
	"github.com/iotaledger/goshimmer/client"
	"github.com/iotaledger/wasp/tools/fairroulette/config"
)

func requestFunds() {
	gosh := client.NewGoShimmerAPI(config.GoshimmerApi())
	_, err := gosh.SendFaucetRequest(Load().Address().String())
	check(err)
}
