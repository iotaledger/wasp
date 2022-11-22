// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package registry

import (
	"fmt"
	"os"
	"path"

	"github.com/iotaledger/hive.go/core/generics/lo"
	"github.com/iotaledger/hive.go/core/generics/onchangemap"
	"github.com/iotaledger/hive.go/core/ioutils"
	"github.com/iotaledger/hive.go/core/kvstore"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/peering"
)

type TrustedPeersRegistry struct {
	storeOnChangeMap *onchangemap.OnChangeMap[string, *peering.ComparablePubKey, *peering.TrustedPeer]
}

var _ TrustedPeersRegistryProvider = &TrustedPeersRegistry{}

// NewTrustedPeersRegistry creates new instance of the trusted peers registry implementation.
func NewTrustedPeersRegistry(storeCallback func(trustedPeers []*peering.TrustedPeer) error) *TrustedPeersRegistry {
	return &TrustedPeersRegistry{
		storeOnChangeMap: onchangemap.NewOnChangeMap[string, *peering.ComparablePubKey](storeCallback),
	}
}

func (p *TrustedPeersRegistry) EnableStoreOnChange() {
	p.storeOnChangeMap.CallbackEnabled(true)
}

func (p *TrustedPeersRegistry) IsTrustedPeer(pubKey *cryptolib.PublicKey) error {
	_, err := p.storeOnChangeMap.Get(peering.NewComparablePubKey(pubKey))
	if err != nil {
		return kvstore.ErrKeyNotFound
	}

	return nil
}

func (p *TrustedPeersRegistry) TrustPeer(pubKey *cryptolib.PublicKey, netID string) (*peering.TrustedPeer, error) {
	trustedPeer := peering.NewTrustedPeer(pubKey, netID)
	if err := p.storeOnChangeMap.Add(trustedPeer); err != nil {
		// already exists, modify the existing
		return p.storeOnChangeMap.Modify(peering.NewComparablePubKey(pubKey), func(item *peering.TrustedPeer) bool {
			*item = *peering.NewTrustedPeer(pubKey, netID)
			return true
		})
	}

	return trustedPeer, nil
}

func (p *TrustedPeersRegistry) DistrustPeer(pubKey *cryptolib.PublicKey) (*peering.TrustedPeer, error) {
	addr := peering.NewComparablePubKey(pubKey)

	trustedPeer, err := p.storeOnChangeMap.Get(addr)
	if err != nil {
		return nil, nil
	}

	if err := p.storeOnChangeMap.Delete(addr); err != nil {
		return nil, nil
	}

	return trustedPeer, nil
}

func (p *TrustedPeersRegistry) TrustedPeers() ([]*peering.TrustedPeer, error) {
	return lo.Values(p.storeOnChangeMap.All()), nil
}

type jsonTrustedPeers struct {
	TrustedPeers []*peering.TrustedPeer `json:"trustedPeers"`
}

func LoadTrustedPeersJSONFromFile(filePath string, trustedPeersRegistry *TrustedPeersRegistry) error {
	tmpTrustedPeers := &jsonTrustedPeers{}
	if err := ioutils.ReadJSONFromFile(filePath, tmpTrustedPeers); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("unable to unmarshal json file: %w", err)
	}

	for _, trustedPeer := range tmpTrustedPeers.TrustedPeers {
		if _, err := trustedPeersRegistry.TrustPeer(trustedPeer.PubKey(), trustedPeer.NetID); err != nil {
			return fmt.Errorf("unable to add trusted peer (%s): %s", filePath, err)
		}
	}

	return nil
}

func WriteTrustedPeersJSONToFile(filePath string, trustedPeers []*peering.TrustedPeer) error {
	if err := os.MkdirAll(path.Dir(filePath), 0o770); err != nil {
		return fmt.Errorf("unable to create folder \"%s\": %w", path.Dir(filePath), err)
	}

	if err := ioutils.WriteJSONToFile(filePath, &jsonTrustedPeers{TrustedPeers: trustedPeers}, 0o600); err != nil {
		return fmt.Errorf("unable to marshal json file: %w", err)
	}

	return nil
}
