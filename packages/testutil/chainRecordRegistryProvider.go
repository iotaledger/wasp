// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package testutil

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/registry_pkg/chainrecord"
)

// Mock implementation of a ChainRecordRegistryProvider for testing purposes

type ChainRecordRegistryProvider struct {
	DB map[coretypes.ChainID]*chainrecord.ChainRecord
}

func NewChainRecordRegistryProvider() *ChainRecordRegistryProvider {
	return &ChainRecordRegistryProvider{
		DB: map[coretypes.ChainID]*chainrecord.ChainRecord{},
	}
}

func (p *ChainRecordRegistryProvider) SaveChainRecord(chainRecord *chainrecord.ChainRecord) error {
	p.DB[*chainRecord.ChainID] = chainRecord
	return nil
}

func (p *ChainRecordRegistryProvider) LoadChainRecord(chainID *coretypes.ChainID) (*chainrecord.ChainRecord, error) {
	ret := p.DB[*chainID]
	return ret, nil
}
