// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package testutil

import (
	"github.com/iotaledger/wasp/v2/packages/registry"
)

// Mock implementation of a ChainRecordRegistryProvider for testing purposes

type ChainRecordRegistryProvider struct {
	rec *registry.ChainRecord
}

func NewChainRecordRegistryProvider() *ChainRecordRegistryProvider {
	return &ChainRecordRegistryProvider{}
}

func (p *ChainRecordRegistryProvider) SaveChainRecord(chainRecord *registry.ChainRecord) error {
	p.rec = chainRecord
	return nil
}

func (p *ChainRecordRegistryProvider) LoadChainRecord() (*registry.ChainRecord, error) {
	return p.rec, nil
}
