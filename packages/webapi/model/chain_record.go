package model

import (
	"github.com/iotaledger/wasp/packages/registry/chainrecord"
)

type ChainRecord struct {
	ChainID ChainID `swagger:"desc(ChainID (base58-encoded))"`
	Active  bool    `swagger:"desc(Whether or not the chain is active)"`
}

func NewChainRecord(rec *chainrecord.ChainRecord) *ChainRecord {
	return &ChainRecord{
		ChainID: NewChainID(rec.ChainID),
		Active:  rec.Active,
	}
}

func (bd *ChainRecord) Record() *chainrecord.ChainRecord {
	return &chainrecord.ChainRecord{
		ChainID: bd.ChainID.ChainID(),
		Active:  bd.Active,
	}
}
