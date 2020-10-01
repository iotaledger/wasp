package dwf

import (
	"github.com/iotaledger/wasp/packages/vm/examples/donatewithfeedback/dwfclient"
	"github.com/iotaledger/wasp/packages/vm/examples/donatewithfeedback/dwfimpl"
	"github.com/iotaledger/wasp/tools/wwallet/config"
	"github.com/iotaledger/wasp/tools/wwallet/sc"
	"github.com/iotaledger/wasp/tools/wwallet/wallet"
)

var Config = &sc.Config{
	ShortName:   "dwf",
	Name:        "DonateWithFeedback",
	ProgramHash: dwfimpl.ProgramHash,
}

func Client() *dwfclient.DWFClient {
	return dwfclient.NewClient(
		config.GoshimmerClient(),
		config.WaspApi(),
		Config.Address(),
		wallet.Load().SignatureScheme(),
	)
}
