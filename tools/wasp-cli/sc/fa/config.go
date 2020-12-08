// +build ignore

package fa

import (
	"github.com/iotaledger/wasp/packages/vm/examples/fairauction"
	"github.com/iotaledger/wasp/packages/vm/examples/fairauction/faclient"
	"github.com/iotaledger/wasp/tools/wasp-cli/sc"
	"github.com/iotaledger/wasp/tools/wasp-cli/wallet"
)

var Config = &sc.Config{
	ShortName:   "fa",
	Name:        "FairAuction",
	ProgramHash: fairauction.ProgramHash,
}

func Client() *faclient.FairAuctionClient {
	return faclient.NewClient(Config.MakeClient(wallet.Load().SignatureScheme()))
}
