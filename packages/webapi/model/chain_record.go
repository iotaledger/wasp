package model

import (
	"github.com/iotaledger/wasp/packages/registry/chainrecord"
)

type ChainRecord struct {
	ChainID ChainID  `swagger:"desc(ChainID (base58-encoded))"`
	Active  bool     `swagger:"desc(Whether or not the chain is active)"`
	Peers   []string `swagger:"desc(List of peers/access nodes (network IDs))"`
}

func NewChainRecord(rec *chainrecord.ChainRecord) *ChainRecord {
	return &ChainRecord{
		ChainID: NewChainID(rec.ChainID),
		Active:  rec.Active,
		Peers:   rec.Peers,
	}
}

func (bd *ChainRecord) Record() *chainrecord.ChainRecord {
	return &chainrecord.ChainRecord{
		ChainID: bd.ChainID.ChainID(),
		Active:  bd.Active,
		Peers:   bd.Peers,
	}
}
