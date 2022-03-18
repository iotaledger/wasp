// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package model

import "github.com/iotaledger/wasp/packages/registry"

type ChainRecord struct {
	ChainID ChainID `swagger:"desc(ChainID (base58-encoded))"`
	Active  bool    `swagger:"desc(Whether or not the chain is active)"`
}

func NewChainRecord(rec *registry.ChainRecord) *ChainRecord {
	return &ChainRecord{
		ChainID: NewChainID(&rec.ChainID),
		Active:  rec.Active,
	}
}

func (bd *ChainRecord) Record() *registry.ChainRecord {
	return &registry.ChainRecord{
		ChainID: *bd.ChainID.ChainID(),
		Active:  bd.Active,
	}
}
