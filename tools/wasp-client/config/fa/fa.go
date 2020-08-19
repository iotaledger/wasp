package fa

import (
	"github.com/iotaledger/wasp/packages/vm/examples/fairauction"
	"github.com/iotaledger/wasp/tools/wasp-client/config"
	"github.com/iotaledger/wasp/tools/wasp-client/wallet"
	"github.com/spf13/pflag"
)

var Config = &config.SCConfig{
	ShortName:   "fa",
	Description: "FairAuction smart contract",
	ProgramHash: fairauction.ProgramHash,
	Flags:       pflag.NewFlagSet("fairauction", pflag.ExitOnError),
}

func Client() *fairauction.FairAuctionClient {
	return fairauction.NewClient(
		config.GoshimmerClient(),
		config.WaspApi(),
		Config.Address(),
		wallet.Load().SignatureScheme(),
	)
}
