package jsonable

import (
	"github.com/iotaledger/wasp/packages/registry"
)

type ChainRecord struct {
	ChainID        *ChainID
	OwnerAddress   *Address
	Color          *Color
	CommitteeNodes []string
	AccessNodes    []string
	Active         bool
}

func NewChainRecord(bd *registry.ChainRecord) *ChainRecord {
	return &ChainRecord{
		ChainID:        NewChainID(&bd.ChainID),
		OwnerAddress:   NewAddress(&bd.OwnerAddress),
		Color:          NewColor(&bd.Color),
		CommitteeNodes: bd.CommitteeNodes[:],
		AccessNodes:    bd.AccessNodes[:],
		Active:         bd.Active,
	}
}

func (bd *ChainRecord) ChainRecord() *registry.ChainRecord {
	return &registry.ChainRecord{
		ChainID:        *bd.ChainID.ChainID(),
		OwnerAddress:   *bd.OwnerAddress.Address(),
		Color:          *bd.Color.Color(),
		CommitteeNodes: bd.CommitteeNodes[:],
		AccessNodes:    bd.AccessNodes[:],
		Active:         bd.Active,
	}
}
