package tr

import (
	"github.com/iotaledger/wasp/packages/vm/examples/tokenregistry"
	"github.com/iotaledger/wasp/packages/vm/examples/tokenregistry/trclient"
	"github.com/iotaledger/wasp/tools/wasp-client/config"
	"github.com/iotaledger/wasp/tools/wasp-client/wallet"
	"github.com/spf13/pflag"
)

var Config = &config.SCConfig{
	ShortName:   "tr",
	Description: "TokenRegistry smart contract",
	ProgramHash: tokenregistry.ProgramHash,
	Flags:       pflag.NewFlagSet("tokenregistry", pflag.ExitOnError),
}

func Client() *trclient.TokenRegistryClient {
	return trclient.NewClient(
		config.GoshimmerClient(),
		config.WaspApi(),
		Config.Address(),
		wallet.Load().SignatureScheme(),
	)
}
