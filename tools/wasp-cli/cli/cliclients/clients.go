// Package cliclients provides client implementations for interacting with various APIs
// within the wasp-cli tool, managing connections and API access.
package cliclients

import (
	"context"

	"github.com/iotaledger/wasp/clients"
	"github.com/iotaledger/wasp/clients/apiclient"
	"github.com/iotaledger/wasp/clients/apiextensions"
	"github.com/iotaledger/wasp/clients/chainclient"
	"github.com/iotaledger/wasp/clients/iota-go/iotaclient"
	"github.com/iotaledger/wasp/components/app"
	"github.com/iotaledger/wasp/packages/isc"
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

func WaspClientWithVersionCheck(ctx context.Context, name string) *apiclient.APIClient {
	client := WaspClientForHostName(name)
	assertMatchingNodeVersion(ctx, name, client)
	return client
}

func assertMatchingNodeVersion(ctx context.Context, name string, client *apiclient.APIClient) {
	if SkipCheckVersions {
		return
	}
	nodeVersion, _, err := client.NodeAPI.
		GetVersion(ctx).
		Execute()
	log.Check(err)
	if app.Version != "v"+nodeVersion.Version {
		// IOTA CLI only warns about a version mismatch, we should do the same. There are rarely real differences between the versions which are relevant.
		log.Printf("node [%s] version: %s, does not match wasp-cli version: %s. You can skip this check by re-running with command with --skip-version-check",
			name, nodeVersion.Version, app.Version)
	}
}

func L2Client() clients.L2Client {
	return L1Client().L2()
}

func L1Client() clients.L1Client {
	return clients.NewL1Client(clients.L1Config{
		APIURL:    config.L1APIAddress(),
		FaucetURL: config.L1FaucetAddress(),
	}, iotaclient.WaitForEffectsEnabled)
}

func ChainClient(waspClient *apiclient.APIClient, chainID isc.ChainID) *chainclient.Client {
	iscPackageID := config.GetPackageID()

	return chainclient.New(
		L1Client(),
		waspClient,
		chainID,
		iscPackageID,
		wallet.Load(),
	)
}
