// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package testutil

import (
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/registry"
)

// Mock implementation of a ChainRecordRegistryProvider for testing purposes

type ChainRecordRegistryProvider struct {
	DB map[isc.ChainID]*registry.ChainRecord
}

func NewChainRecordRegistryProvider() *ChainRecordRegistryProvider {
	return &ChainRecordRegistryProvider{
		DB: map[isc.ChainID]*registry.ChainRecord{},
	}
}

func (p *ChainRecordRegistryProvider) AddChainRecord(chainRecord *registry.ChainRecord) error {
	p.DB[chainRecord.ChainID()] = chainRecord
	return nil
}

func (p *ChainRecordRegistryProvider) LoadChainRecord() (*registry.ChainRecord, error) {
	ret := p.DB[chainID]
	return ret, nil
}
