package chain

import (
	"github.com/iotaledger/wasp/client/chainclient"
	"github.com/iotaledger/wasp/client/multiclient"
	"github.com/iotaledger/wasp/client/scclient"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/tools/wasp-cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/wallet"
)

func Client() *chainclient.Client {
	return chainclient.New(
		config.GoshimmerClient(),
		config.WaspClient(),
		*GetCurrentChainID(),
		wallet.Load().KeyPair(),
	)
}

func MultiClient() *multiclient.MultiClient {
	return multiclient.New(config.CommitteeApi(chainCommittee()))
}

func SCClient(contractHname coretypes.Hname) *scclient.SCClient {
	return scclient.New(Client(), contractHname)
}
