package model

import (
	"github.com/iotaledger/wasp/packages/registry_pkg"
)

type ChainRecord struct {
	ChainID ChainID `swagger:"desc(ChainID (base58-encoded))"`
	Active  bool    `swagger:"desc(Whether or not the chain is active)"`
}

func NewChainRecord(rec *registry_pkg.ChainRecord) *ChainRecord {
	return &ChainRecord{
		ChainID: NewChainID(rec.ChainID),
		Active:  rec.Active,
	}
}

func (bd *ChainRecord) Record() *registry_pkg.ChainRecord {
	return &registry_pkg.ChainRecord{
		ChainID: bd.ChainID.ChainID(),
		Active:  bd.Active,
	}
}
