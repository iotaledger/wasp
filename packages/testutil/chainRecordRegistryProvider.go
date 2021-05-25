// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package testutil

import (
	"github.com/iotaledger/wasp/packages/coretypes"
)

// Mock implementation of a ChainRecordRegistryProvider for testing purposes

type ChainRecordRegistryProvider struct {
	DB map[coretypes.ChainID]*coretypes.ChainRecord
}

func NewChainRecordRegistryProvider() *ChainRecordRegistryProvider {
	return &ChainRecordRegistryProvider{
		DB: map[coretypes.ChainID]*coretypes.ChainRecord{},
	}
}

func (p *ChainRecordRegistryProvider) SaveChainRecord(chainRecord *coretypes.ChainRecord) error {
	p.DB[*chainRecord.ChainID] = chainRecord
	return nil
}

func (p *ChainRecordRegistryProvider) LoadChainRecord(chainID *coretypes.ChainID) (*coretypes.ChainRecord, error) {
	ret := p.DB[*chainID]
	return ret, nil
}
