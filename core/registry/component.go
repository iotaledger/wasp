// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package registry

import (
	"os"
	"path/filepath"

	"go.uber.org/dig"

	"github.com/iotaledger/hive.go/core/app"
	hivep2p "github.com/iotaledger/hive.go/core/p2p"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/tcrypto"
)

func init() {
	CoreComponent = &app.CoreComponent{
		Component: &app.Component{
			Name:    "Registry",
			Params:  params,
			Provide: provide,
		},
	}
}

var CoreComponent *app.CoreComponent

//nolint:funlen
func provide(c *dig.Container) error {
	type nodeIdentityProviderResult struct {
		dig.Out
		NodeIdentityProvider registry.NodeIdentityProvider
	}

	if err := c.Provide(func() nodeIdentityProviderResult {
		if err := os.MkdirAll(ParamsP2P.Database.Path, 0o700); err != nil {
			CoreComponent.LogPanicf("could not create peer store database dir '%s': %w", ParamsP2P.Database.Path, err)
		}

		privKeyFilePath := filepath.Join(ParamsP2P.Database.Path, "identity.key")

		// make sure nobody copies around the peer store since it contains the private key of the node
		CoreComponent.LogInfof(`WARNING: never share your "%s" folder as it contains your node's private key!`, ParamsP2P.Database.Path)

		// load up the previously generated identity or create a new one
		privKey, newlyCreated, err := hivep2p.LoadOrCreateIdentityPrivateKey(privKeyFilePath, ParamsP2P.IdentityPrivateKey)
		if err != nil {
			CoreComponent.LogPanic(err)
		}

		if newlyCreated {
			CoreComponent.LogInfof(`stored new private key for peer identity under "%s"`, privKeyFilePath)
		} else {
			CoreComponent.LogInfof(`loaded existing private key for peer identity from "%s"`, privKeyFilePath)
		}

		privKeyBytes, err := privKey.Raw()
		if err != nil {
			CoreComponent.LogPanicf("unable to convert private key for peer identity: %s", err)
		}

		waspPrivKey, err := cryptolib.NewPrivateKeyFromBytes(privKeyBytes)
		if err != nil {
			CoreComponent.LogPanicf("unable to convert private key for peer identity: %s", err)
		}

		return nodeIdentityProviderResult{
			NodeIdentityProvider: registry.NewNodeIdentity(cryptolib.NewKeyPairFromPrivateKey(waspPrivKey)),
		}
	}); err != nil {
		CoreComponent.LogPanic(err)
	}

	type chainRegistryResult struct {
		dig.Out
		ChainRecordRegistryProvider registry.ChainRecordRegistryProvider
	}

	if err := c.Provide(func() chainRegistryResult {
		filePath := ParamsRegistry.Chains.FilePath

		chainRecordRegistry := registry.NewChainRecordRegistry(
			func(chainRecords []*registry.ChainRecord) error {
				return registry.WriteChainRecordsJSONToFile(filePath, chainRecords)
			},
		)

		// load chain records on startup
		if err := registry.LoadChainRecordsJSONFromFile(filePath, chainRecordRegistry); err != nil {
			CoreComponent.LogPanicf("unable to read chain records configuration (%s): %s", filePath, err)
		}

		chainRecordRegistry.EnableStoreOnChange()

		return chainRegistryResult{
			ChainRecordRegistryProvider: chainRecordRegistry,
		}
	}); err != nil {
		CoreComponent.LogPanic(err)
	}

	type dkSharesRegistryDeps struct {
		dig.In
		NodeIdentityProvider registry.NodeIdentityProvider
	}

	type dkSharesRegistryResult struct {
		dig.Out
		DKShareRegistryProvider registry.DKShareRegistryProvider
	}

	if err := c.Provide(func(deps dkSharesRegistryDeps) dkSharesRegistryResult {
		filePath := ParamsRegistry.DKShares.FilePath

		dkSharesRegistry := registry.NewDKSharesRegistry(
			func(dkShares []tcrypto.DKShare) error {
				return registry.WriteDKSharesJSONToFile(filePath, dkShares)
			},
		)

		// load DKShares on startup
		if err := registry.LoadDKSharesJSONFromFile(filePath, dkSharesRegistry, deps.NodeIdentityProvider.NodeIdentity().GetPrivateKey()); err != nil {
			CoreComponent.LogPanicf("unable to read DKShares configuration (%s): %s", filePath, err)
		}

		dkSharesRegistry.EnableStoreOnChange()

		return dkSharesRegistryResult{
			DKShareRegistryProvider: dkSharesRegistry,
		}
	}); err != nil {
		CoreComponent.LogPanic(err)
	}

	type trustedPeersRegistryResult struct {
		dig.Out
		TrustedPeersRegistryProvider registry.TrustedPeersRegistryProvider `name:"TrustedPeersRegistryProvider"`
	}

	if err := c.Provide(func() trustedPeersRegistryResult {
		filePath := ParamsRegistry.TrustedPeers.FilePath

		trustedPeersRegistry := registry.NewTrustedPeersRegistry(
			func(trustedPeers []*peering.TrustedPeer) error {
				return registry.WriteTrustedPeersJSONToFile(filePath, trustedPeers)
			},
		)

		// load TrustedPeers on startup
		if err := registry.LoadTrustedPeersJSONFromFile(filePath, trustedPeersRegistry); err != nil {
			CoreComponent.LogPanicf("unable to read TrustedPeers configuration (%s): %s", filePath, err)
		}

		trustedPeersRegistry.EnableStoreOnChange()

		return trustedPeersRegistryResult{
			TrustedPeersRegistryProvider: trustedPeersRegistry,
		}
	}); err != nil {
		CoreComponent.LogPanic(err)
	}

	return nil
}
