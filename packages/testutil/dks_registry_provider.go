// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package testutil

import (
	"fmt"

	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/registry"
	"github.com/iotaledger/wasp/v2/packages/tcrypto"
)

// DkgRegistry stands for a mock for dkg.DKShareRegistry.
type DkgRegistry struct {
	DB          map[string][]byte
	nodePrivKey *cryptolib.PrivateKey
}

var _ registry.DKShareRegistry = &DkgRegistry{}

// NewDkgRegistry creates new mocked DKG registry.
func NewDkgRegistry(nodePrivKey *cryptolib.PrivateKey) *DkgRegistry {
	return &DkgRegistry{
		DB:          map[string][]byte{},
		nodePrivKey: nodePrivKey,
	}
}

// SaveDKShare implements dkg.DKShareRegistry.
func (p *DkgRegistry) SaveDKShare(dkShare tcrypto.DKShare) error {
	p.DB[dkShare.GetAddress().String()] = dkShare.Bytes()
	return nil
}

// LoadDKShare implements dkg.DKShareRegistry.
func (p *DkgRegistry) LoadDKShare(sharedAddress *cryptolib.Address) (tcrypto.DKShare, error) {
	dkShareBytes := p.DB[sharedAddress.String()]
	if dkShareBytes == nil {
		return nil, fmt.Errorf("DKShare not found for %v", sharedAddress.String())
	}
	return tcrypto.DKShareFromBytes(dkShareBytes, tcrypto.DefaultEd25519Suite(), tcrypto.DefaultBLSSuite(), p.nodePrivKey)
}
