package fr

import (
	"github.com/iotaledger/wasp/packages/vm/examples/fairroulette"
	"github.com/iotaledger/wasp/packages/vm/examples/fairroulette/frclient"
	"github.com/iotaledger/wasp/tools/wasp-cli/sc"
	"github.com/iotaledger/wasp/tools/wasp-cli/wallet"
)

var Config = &sc.Config{
	ShortName:   "fr",
	Name:        "FairRoulette",
	ProgramHash: fairroulette.ProgramHash,
}

func Client() *frclient.FairRouletteClient {
	return frclient.NewClient(Config.MakeClient(wallet.Load().SignatureScheme()))
}
