// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package testutil

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/registry_pkg/chain_record"
)

// Mock implementation of a ChainRecordRegistryProvider for testing purposes

type ChainRecordRegistryProvider struct {
	DB map[coretypes.ChainID]*chain_record.ChainRecord
}

func NewChainRecordRegistryProvider() *ChainRecordRegistryProvider {
	return &ChainRecordRegistryProvider{
		DB: map[coretypes.ChainID]*chain_record.ChainRecord{},
	}
}

func (p *ChainRecordRegistryProvider) SaveChainRecord(chainRecord *chain_record.ChainRecord) error {
	p.DB[*coretypes.NewChainID(chainRecord.ChainIdAliasAddress)] = chainRecord
	return nil
}

func (p *ChainRecordRegistryProvider) LoadChainRecord(chainID *coretypes.ChainID) (*chain_record.ChainRecord, error) {
	ret := p.DB[*chainID]
	return ret, nil
}
