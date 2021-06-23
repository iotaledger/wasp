// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package testutil

import (
	"fmt"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"

	"github.com/iotaledger/wasp/packages/tcrypto"
)

// DkgRegistryProvider stands for a mock for dkg.DKShareRegistryProvider.
type DkgRegistryProvider struct {
	DB    map[string][]byte
	Suite tcrypto.Suite
}

// NewDkgRegistryProvider creates new mocked DKG registry provider.
func NewDkgRegistryProvider(suite tcrypto.Suite) *DkgRegistryProvider { // TODO: XXX: KP: Remove the SUITE.
	return &DkgRegistryProvider{
		DB:    map[string][]byte{},
		Suite: suite,
	}
}

// SaveDKShare implements dkg.DKShareRegistryProvider.
func (p *DkgRegistryProvider) SaveDKShare(dkShare *tcrypto.DKShare) error {
	var err error
	var dkShareBytes []byte
	if dkShareBytes, err = dkShare.Bytes(); err != nil {
		return err
	}
	p.DB[dkShare.Address.String()] = dkShareBytes
	return nil
}

// LoadDKShare implements dkg.DKShareRegistryProvider.
func (p *DkgRegistryProvider) LoadDKShare(sharedAddress ledgerstate.Address) (*tcrypto.DKShare, error) {
	var dkShareBytes = p.DB[sharedAddress.String()]
	if dkShareBytes == nil {
		return nil, fmt.Errorf("DKShare not found for %v", sharedAddress.Base58())
	}
	return tcrypto.DKShareFromBytes(dkShareBytes, p.Suite)
}
