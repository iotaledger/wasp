package chain

import (
	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/client/scclient"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/tools/wwallet/config"
	"github.com/iotaledger/wasp/tools/wwallet/wallet"
)

func Client() *chainclient.Client {
	return chainclient.New(
		config.GoshimmerClient(),
		config.WaspClient(),
		GetCurrentChainID(),
		wallet.Load().SignatureScheme(),
	)
}

func SCClient(contractHname coretypes.Hname) *scclient.SCClient {
	return scclient.New(Client(), contractHname)
}
