package model

import (
	"github.com/iotaledger/wasp/packages/registry"
)

type ChainRecord struct {
	ChainID        ChainID  `swagger:"desc(ChainID (base58-encoded))"`
	Color          Color    `swagger:"desc(Chain color (base58-encoded))"`
	CommitteeNodes []string `swagger:"desc(List of committee nodes (network IDs))"`
	Active         bool     `swagger:"desc(Whether or not the chain is active)"`
}

func NewChainRecord(bd *registry.ChainRecord) *ChainRecord {
	return &ChainRecord{
		ChainID:        NewChainID(&bd.ChainID),
		Color:          NewColor(&bd.Color),
		CommitteeNodes: bd.CommitteeNodes[:],
		Active:         bd.Active,
	}
}

func (bd *ChainRecord) ChainRecord() *registry.ChainRecord {
	return &registry.ChainRecord{
		ChainID:        bd.ChainID.ChainID(),
		Color:          bd.Color.Color(),
		CommitteeNodes: bd.CommitteeNodes[:],
		Active:         bd.Active,
	}
}
