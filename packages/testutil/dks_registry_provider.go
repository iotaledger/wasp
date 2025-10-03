// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package testutil

import (
	"fmt"

	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/registry"
	"github.com/iotaledger/wasp/v2/packages/tcrypto"
)

// DkgRegistryProvider stands for a mock for distKeyGen.DistKeyPartRegistryProvider.
type DkgRegistryProvider struct {
	DB          map[string][]byte
	nodePrivKey *cryptolib.PrivateKey
}

var _ registry.DistKeyPartRegistryProvider = &DkgRegistryProvider{}

// NewDkgRegistryProvider creates new mocked DistKeyGeneration registry provider.
func NewDkgRegistryProvider(nodePrivKey *cryptolib.PrivateKey) *DkgRegistryProvider {
	return &DkgRegistryProvider{
		DB:          map[string][]byte{},
		nodePrivKey: nodePrivKey,
	}
}

// SaveDistKeyPart implements distKeyGen.DistKeyPartRegistryProvider.
func (p *DkgRegistryProvider) SaveDistKeyPart(distKeyPart tcrypto.DistibutedKeyPart) error {
	p.DB[distKeyPart.GetAddress().String()] = distKeyPart.Bytes()
	return nil
}

// LoadDistKeyPart implements distKeyGen.DistKeyPartRegistryProvider.
func (p *DkgRegistryProvider) LoadDistKeyPart(sharedAddress *cryptolib.Address) (tcrypto.DistibutedKeyPart, error) {
	distKeyPartBytes := p.DB[sharedAddress.String()]
	if distKeyPartBytes == nil {
		return nil, fmt.Errorf("DistKeyPart not found for %v", sharedAddress.String())
	}
	return tcrypto.DistKeyPartFromBytes(distKeyPartBytes, tcrypto.DefaultEd25519Suite(), tcrypto.DefaultBLSSuite(), p.nodePrivKey)
}
