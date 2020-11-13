package jsonable

import (
	"github.com/iotaledger/wasp/packages/registry"
)

type ChainRecord struct {
	ChainID        *ChainID
	Color          *Color
	CommitteeNodes []string
	Active         bool
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
		ChainID:        *bd.ChainID.ChainID(),
		Color:          *bd.Color.Color(),
		CommitteeNodes: bd.CommitteeNodes[:],
		Active:         bd.Active,
	}
}
