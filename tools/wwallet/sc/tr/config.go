package tr

import (
	"github.com/iotaledger/wasp/packages/vm/examples/tokenregistry"
	"github.com/iotaledger/wasp/packages/vm/examples/tokenregistry/trclient"
	"github.com/iotaledger/wasp/tools/wwallet/config"
	"github.com/iotaledger/wasp/tools/wwallet/sc"
	"github.com/iotaledger/wasp/tools/wwallet/wallet"
)

var Config = &sc.Config{
	ShortName:   "tr",
	Name:        "TokenRegistry",
	ProgramHash: tokenregistry.ProgramHash,
}

func Client() *trclient.TokenRegistryClient {
	return trclient.NewClient(
		config.GoshimmerClient(),
		config.WaspApi(),
		Config.Address(),
		wallet.Load().SignatureScheme(),
	)
}
