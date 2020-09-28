package fa

import (
	"github.com/iotaledger/wasp/packages/vm/examples/fairauction"
	"github.com/iotaledger/wasp/packages/vm/examples/fairauction/faclient"
	"github.com/iotaledger/wasp/tools/wwallet/config"
	"github.com/iotaledger/wasp/tools/wwallet/sc"
	"github.com/iotaledger/wasp/tools/wwallet/wallet"
	"github.com/spf13/pflag"
)

var Config = &sc.Config{
	ShortName:   "fa",
	Name:        "FairAuction",
	ProgramHash: fairauction.ProgramHash,
	Flags:       pflag.NewFlagSet("FairAuction", pflag.ExitOnError),
}

func Client() *faclient.FairAuctionClient {
	return faclient.NewClient(
		config.GoshimmerClient(),
		config.WaspApi(),
		Config.Address(),
		wallet.Load().SignatureScheme(),
	)
}
