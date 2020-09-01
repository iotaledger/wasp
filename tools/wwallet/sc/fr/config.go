package fr

import (
	"github.com/iotaledger/wasp/packages/vm/examples/fairroulette"
	"github.com/iotaledger/wasp/packages/vm/examples/fairroulette/frclient"
	"github.com/iotaledger/wasp/tools/wwallet/config"
	"github.com/iotaledger/wasp/tools/wwallet/sc"
	"github.com/iotaledger/wasp/tools/wwallet/wallet"
	"github.com/spf13/pflag"
)

var Config = &sc.Config{
	ShortName:   "fr",
	Description: "FairRoulette smart contract",
	ProgramHash: fairroulette.ProgramHash,
	Flags:       pflag.NewFlagSet("fairroulette", pflag.ExitOnError),
}

func Client() *frclient.FairRouletteClient {
	return frclient.NewClient(
		config.GoshimmerClient(),
		config.WaspApi(),
		Config.Address(),
		wallet.Load().SignatureScheme(),
	)
}
