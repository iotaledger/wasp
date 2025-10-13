// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package testutil

import (
	"fmt"

	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/registry"
	"github.com/iotaledger/wasp/v2/packages/tcrypto"
)

// DistributedKeyGenerationRegistryProvider stands for a mock for dkg.DKShareRegistryProvider.
type DistributedKeyGenerationRegistryProvider struct {
	DB          map[string][]byte
	nodePrivKey *cryptolib.PrivateKey
}

var _ registry.DKShareRegistryProvider = &DistributedKeyGenerationRegistryProvider{}

// NewDistributedKeyGenerationRegistryProvider creates new mocked DKG registry provider.
func NewDistributedKeyGenerationRegistryProvider(nodePrivKey *cryptolib.PrivateKey) *DistributedKeyGenerationRegistryProvider {
	return &DistributedKeyGenerationRegistryProvider{
		DB:          map[string][]byte{},
		nodePrivKey: nodePrivKey,
	}
}

// SaveDKShare implements dkg.DKShareRegistryProvider.
func (p *DistributedKeyGenerationRegistryProvider) SaveDKShare(dkShare tcrypto.DKShare) error {
	p.DB[dkShare.GetAddress().String()] = dkShare.Bytes()
	return nil
}

// LoadDKShare implements dkg.DKShareRegistryProvider.
func (p *DistributedKeyGenerationRegistryProvider) LoadDKShare(sharedAddress *cryptolib.Address) (tcrypto.DKShare, error) {
	dkShareBytes := p.DB[sharedAddress.String()]
	if dkShareBytes == nil {
		return nil, fmt.Errorf("DKShare not found for %v", sharedAddress.String())
	}
	return tcrypto.DKShareFromBytes(dkShareBytes, tcrypto.DefaultEd25519Suite(), tcrypto.DefaultBLSSuite(), p.nodePrivKey)
}
