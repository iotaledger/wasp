package clients

import (
	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/packages/l1connection"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

func WaspClientForNodeNumber(i ...int) *apiclient.APIClient {
	return WaspClient(config.WaspAPIURL(i...))
}

func WaspClient(apiAddress string) *apiclient.APIClient {
	// TODO: add authentication for /adm
	L1Client() // this will fill parameters.L1() with data from the L1 node
	log.Verbosef("using Wasp host %s\n", apiAddress)

	apiConfig := apiclient.NewConfiguration()
	apiConfig.Host = apiAddress
	apiConfig.AddDefaultHeader("Authorization", config.GetToken())

	return apiclient.NewAPIClient(apiConfig)
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
