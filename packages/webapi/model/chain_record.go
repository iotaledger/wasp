// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package model

import "github.com/iotaledger/wasp/packages/registry"

type ChainRecord struct {
	ChainID ChainIDBech32 `swagger:"desc(ChainID (bech32))"`
	Active  bool          `swagger:"desc(Whether or not the chain is active)"`
}

func NewChainRecord(rec *registry.ChainRecord) *ChainRecord {
	chainID := rec.ChainID()
	return &ChainRecord{
		ChainID: NewChainIDBech32(chainID),
		Active:  rec.Active,
	}
}

func (bd *ChainRecord) Record() *registry.ChainRecord {
	return registry.NewChainRecord(bd.ChainID.ChainID(), bd.Active)
}
