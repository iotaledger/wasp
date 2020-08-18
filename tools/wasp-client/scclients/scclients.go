package scclients

import (
	"github.com/iotaledger/wasp/packages/vm/examples/fairroulette"
	"github.com/iotaledger/wasp/tools/wasp-client/config"
	"github.com/iotaledger/wasp/tools/wasp-client/wallet"
)

func GetFRClient() *fairroulette.FairRouletteClient {
	scAddress := config.GetFRAddress()
	return fairroulette.NewClient(
		config.GoshimmerClient(),
		config.WaspApi(),
		&scAddress,
		wallet.Load().SignatureScheme(),
	)
}
