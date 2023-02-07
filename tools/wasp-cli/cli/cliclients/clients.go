package cliclients

import (
	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/clients/apiextensions"
	"github.com/iotaledger/wasp/clients/chainclient"
	"github.com/iotaledger/wasp/clients/scclient"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/l1connection"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/wallet"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

func WaspClientForHostName(apiAddress string) *apiclient.APIClient {
	L1Client() // this will fill parameters.L1() with data from the L1 node
	log.Verbosef("using Wasp host %s\n", apiAddress)

	client, err := apiextensions.WaspAPIClientByHostName(apiAddress)
	log.Check(err)

	client.GetConfig().AddDefaultHeader("Authorization", "Bearer "+config.GetToken())

	return client
}

func WaspClientForIndex(i ...int) *apiclient.APIClient {
	return WaspClientForHostName(config.MustWaspAPIURL(i...))
}

func L1Client() l1connection.Client {
	log.Verbosef("using L1 API %s\n", config.L1APIAddress())

	return l1connection.NewClient(
		l1connection.Config{
			APIAddress:    config.L1APIAddress(),
			FaucetAddress: config.L1FaucetAddress(),
		},
		log.HiveLogger(),
	)
}

func ChainClient(waspClient *apiclient.APIClient) *chainclient.Client {
	return chainclient.New(
		L1Client(),
		waspClient,
		config.GetCurrentChainID(),
		wallet.Load().KeyPair,
	)
}

func SCClient(apiClient *apiclient.APIClient, contractHname isc.Hname) *scclient.SCClient {
	return scclient.New(ChainClient(apiClient), contractHname)
}
