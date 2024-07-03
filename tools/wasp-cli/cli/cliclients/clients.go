package cliclients

import (
	"context"

	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/clients/apiextensions"
	"github.com/iotaledger/wasp/clients/chainclient"
	"github.com/iotaledger/wasp/clients/iscmove"
	"github.com/iotaledger/wasp/components/app"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/l2connection"
	"github.com/iotaledger/wasp/sui-go/suiclient"
	"github.com/iotaledger/wasp/sui-go/suiconn"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/config"
	"github.com/iotaledger/wasp/tools/wasp-cli/cli/wallet"
	"github.com/iotaledger/wasp/tools/wasp-cli/log"
)

var SkipCheckVersions bool

func WaspClientForHostName(name string) *apiclient.APIClient {
	apiAddress := config.MustWaspAPIURL(name)
	L2Client() // this will fill parameters.L1() with data from the L1 node
	log.Verbosef("using Wasp host %s\n", apiAddress)

	client, err := apiextensions.WaspAPIClientByHostName(apiAddress)
	log.Check(err)

	client.GetConfig().Debug = log.DebugFlag
	client.GetConfig().AddDefaultHeader("Authorization", "Bearer "+config.GetToken(name))

	return client
}

func WaspClient(name string) *apiclient.APIClient {
	client := WaspClientForHostName(name)
	assertMatchingNodeVersion(name, client)
	return client
}

func assertMatchingNodeVersion(name string, client *apiclient.APIClient) {
	if SkipCheckVersions {
		return
	}
	nodeVersion, _, err := client.NodeApi.
		GetVersion(context.Background()).
		Execute()
	log.Check(err)
	if app.Version != "v"+nodeVersion.Version {
		log.Fatalf("node [%s] version: %s, does not match wasp-cli version: %s. You can skip this check by re-running with command with --skip-version-check",
			name, nodeVersion.Version, app.Version)
	}
}

func L2Client() l2connection.Client {
	log.Verbosef("using L1 API %s\n", config.L1APIAddress())

	return iscmove.NewClient(
		iscmove.Config{
			APIURL:       suiconn.LocalnetEndpointURL,
			FaucetURL:    suiconn.LocalnetFaucetURL,
			WebsocketURL: suiconn.LocalnetWebsocketEndpointURL,
		},
	)
}

func L1Client() *suiclient.Client {
	return suiclient.New(suiconn.LocalnetEndpointURL)
}

func ChainClient(waspClient *apiclient.APIClient, chainID isc.ChainID) *chainclient.Client {
	return chainclient.New(
		L2Client(),
		waspClient,
		chainID,
		wallet.Load(),
	)
}
