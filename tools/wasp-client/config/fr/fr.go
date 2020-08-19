package fr

import (
	"github.com/iotaledger/wasp/packages/vm/examples/fairroulette"
	"github.com/iotaledger/wasp/tools/wasp-client/config"
	"github.com/iotaledger/wasp/tools/wasp-client/wallet"
	"github.com/spf13/pflag"
)

var Config = &config.SCConfig{
	ShortName: "fr",
	Flags:     pflag.NewFlagSet("fairroulette", pflag.ExitOnError),
}

func Client() *fairroulette.FairRouletteClient {
	return fairroulette.NewClient(
		config.GoshimmerClient(),
		config.WaspApi(),
		Config.Address(),
		wallet.Load().SignatureScheme(),
	)
}
