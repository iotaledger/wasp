// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package registry

import (
	"context"
	"fmt"
	"os"

	"github.com/iotaledger/hive.go/core/generics/event"
	"github.com/iotaledger/hive.go/core/generics/lo"
	"github.com/iotaledger/hive.go/core/generics/onchangemap"
	"github.com/iotaledger/hive.go/core/ioutils"
	"github.com/iotaledger/hive.go/core/kvstore"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/util"
)

type jsonTrustedPeers struct {
	TrustedPeers []*peering.TrustedPeer `json:"trustedPeers"`
}

type TrustedPeersRegistryImpl struct {
	onChangeMap  *onchangemap.OnChangeMap[string, *peering.ComparablePubKey, *peering.TrustedPeer]
	changeEvents *event.Event[[]*peering.TrustedPeer]

	filePath string
}

var _ TrustedPeersRegistryProvider = &TrustedPeersRegistryImpl{}

// NewTrustedPeersRegistryImpl creates new instance of the trusted peers registry implementation.
func NewTrustedPeersRegistryImpl(filePath string) (*TrustedPeersRegistryImpl, error) {
	// create the target directory during initialization
	if err := util.CreateDirectoryForFilePath(filePath, 0o770); err != nil {
		return nil, err
	}

	registry := &TrustedPeersRegistryImpl{
		filePath:     filePath,
		changeEvents: event.New[[]*peering.TrustedPeer](),
	}

	registry.onChangeMap = onchangemap.NewOnChangeMap(
		onchangemap.WithChangedCallback[string, *peering.ComparablePubKey](registry.trustedPeersUpdated),
	)

	// load TrustedPeers on startup
	if err := registry.loadTrustedPeersJSON(); err != nil {
		return nil, fmt.Errorf("unable to read TrustedPeers configuration (%s): %w", filePath, err)
	}

	registry.onChangeMap.CallbacksEnabled(true)

	return registry, nil
}

func (p *TrustedPeersRegistryImpl) loadTrustedPeersJSON() error {
	if p.filePath == "" {
		// do not load entries if no path is given
		return nil
	}

	tmpTrustedPeers := &jsonTrustedPeers{}
	if err := ioutils.ReadJSONFromFile(p.filePath, tmpTrustedPeers); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("unable to unmarshal json file: %w", err)
	}

	for _, trustedPeer := range tmpTrustedPeers.TrustedPeers {
		if _, err := p.TrustPeer(trustedPeer.Name, trustedPeer.PubKey(), trustedPeer.PeeringURL); err != nil {
			return fmt.Errorf("unable to add trusted peer (%s): %w", p.filePath, err)
		}
	}

	return nil
}

func (p *TrustedPeersRegistryImpl) trustedPeersUpdated(trustedPeers []*peering.TrustedPeer) error {
	p.changeEvents.Trigger(trustedPeers)
	return p.writeTrustedPeersJSON(trustedPeers)
}

func (p *TrustedPeersRegistryImpl) writeTrustedPeersJSON(trustedPeers []*peering.TrustedPeer) error {
	if p.filePath == "" {
		// do not store entries if no path is given
		return nil
	}

	if err := util.CreateDirectoryForFilePath(p.filePath, 0o770); err != nil {
		return err
	}

	if err := ioutils.WriteJSONToFile(p.filePath, &jsonTrustedPeers{TrustedPeers: trustedPeers}, 0o600); err != nil {
		return fmt.Errorf("unable to marshal json file: %w", err)
	}

	return nil
}

func (p *TrustedPeersRegistryImpl) IsTrustedPeer(pubKey *cryptolib.PublicKey) error {
	_, err := p.onChangeMap.Get(peering.NewComparablePubKey(pubKey))
	if err != nil {
		return kvstore.ErrKeyNotFound
	}

	return nil
}

func (p *TrustedPeersRegistryImpl) TrustPeer(name string, pubKey *cryptolib.PublicKey, peeringURL string) (*peering.TrustedPeer, error) {
	if err := peering.ValidateTrustedPeerParams(name, pubKey, peeringURL); err != nil {
		return nil, err
	}
	for _, existingPeer := range p.onChangeMap.All() {
		if existingPeer.Name == name && !existingPeer.PubKey().Equals(pubKey) {
			return nil, fmt.Errorf("peer with name \"%s\" already exists", name)
		}
	}
	trustedPeer := peering.NewTrustedPeer(name, pubKey, peeringURL)
	if err := p.onChangeMap.Add(trustedPeer); err != nil {
		// already exists, modify the existing
		return p.onChangeMap.Modify(peering.NewComparablePubKey(pubKey), func(item *peering.TrustedPeer) bool {
			*item = *peering.NewTrustedPeer(name, pubKey, peeringURL)
			return true
		})
	}

	return trustedPeer, nil
}

func (p *TrustedPeersRegistryImpl) DistrustPeer(pubKey *cryptolib.PublicKey) (*peering.TrustedPeer, error) {
	addr := peering.NewComparablePubKey(pubKey)

	trustedPeer, err := p.onChangeMap.Get(addr)
	if err != nil {
		return nil, nil
	}

	if err := p.onChangeMap.Delete(addr); err != nil {
		return nil, nil
	}

	return trustedPeer, nil
}

func (p *TrustedPeersRegistryImpl) TrustedPeers() ([]*peering.TrustedPeer, error) {
	return lo.Values(p.onChangeMap.All()), nil
}

func (p *TrustedPeersRegistryImpl) TrustedPeersByPubKeyOrName(pubKeysOrNames []string) ([]*peering.TrustedPeer, error) {
	return peering.QueryByPubKeyOrName(lo.Values(p.onChangeMap.All()), pubKeysOrNames)
}

func (p *TrustedPeersRegistryImpl) mustTrustedPeers() []*peering.TrustedPeer {
	return lo.Values(p.onChangeMap.All())
}

func (p *TrustedPeersRegistryImpl) TrustedPeersListener(callback func([]*peering.TrustedPeer)) context.CancelFunc {
	callback(p.mustTrustedPeers())
	closure := event.NewClosure(callback)
	p.changeEvents.Attach(closure)
	return func() { p.changeEvents.Detach(closure) }
}
