// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package testutil

import (
	"fmt"

	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/registry"
	"github.com/iotaledger/wasp/v2/packages/tcrypto"
)

// DistributedKeyGenerationRegistry stands for a mock for distkeygen.DistKeyPartsRegistry.
type DistributedKeyGenerationRegistry struct {
	DB          map[string][]byte
	nodePrivKey *cryptolib.PrivateKey
}

var _ registry.DistKeyPartsRegistry = &DistributedKeyGenerationRegistry{}

// NewDistributedKeyGenerationRegistry creates new mocked DKG registry provider.
func NewDistributedKeyGenerationRegistry(nodePrivKey *cryptolib.PrivateKey) *DistributedKeyGenerationRegistry {
	return &DistributedKeyGenerationRegistry{
		DB:          map[string][]byte{},
		nodePrivKey: nodePrivKey,
	}
}

// SaveDistKeyPart implements distkeygen.DistKeyPartsRegistry.
func (p *DistributedKeyGenerationRegistry) SaveDistKeyPart(distKeyPart tcrypto.DistKeyPart) error {
	p.DB[distKeyPart.GetAddress().String()] = distKeyPart.Bytes()
	return nil
}

// LoadDistKeyPart implements distkeygen.DistKeyPartsRegistry.
func (p *DistributedKeyGenerationRegistry) LoadDistKeyPart(sharedAddress *cryptolib.Address) (tcrypto.DistKeyPart, error) {
	distKeyPartBytes := p.DB[sharedAddress.String()]
	if distKeyPartBytes == nil {
		return nil, fmt.Errorf("DistKeyPart not found for %v", sharedAddress.String())
	}
	return tcrypto.DistKeyPartFromBytes(distKeyPartBytes, tcrypto.DefaultEd25519Suite(), tcrypto.DefaultBLSSuite(), p.nodePrivKey)
}
