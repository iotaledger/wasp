package model

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/registry/chainrecord"
)

type ChainRecord struct {
	ChainID ChainID `swagger:"desc(ChainID (base58-encoded))"`
	Active  bool    `swagger:"desc(Whether or not the chain is active)"`
}

func NewChainRecord(rec *chainrecord.ChainRecord) *ChainRecord {
	return &ChainRecord{
		ChainID: NewChainID(coretypes.NewChainID(rec.ChainAddr)),
		Active:  rec.Active,
	}
}

func (bd *ChainRecord) Record() *chainrecord.ChainRecord {
	return &chainrecord.ChainRecord{
		ChainAddr: bd.ChainID.ChainID().AliasAddress,
		Active:    bd.Active,
	}
}
