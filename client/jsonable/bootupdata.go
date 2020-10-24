package jsonable

import (
	"github.com/iotaledger/wasp/packages/registry"
)

type BootupData struct {
	ChainID        *ChainID
	OwnerAddress   *Address
	Color          *Color
	CommitteeNodes []string
	AccessNodes    []string
	Active         bool
}

func NewBootupData(bd *registry.BootupData) *BootupData {
	return &BootupData{
		ChainID:        NewChainID(&bd.ChainID),
		OwnerAddress:   NewAddress(&bd.OwnerAddress),
		Color:          NewColor(&bd.Color),
		CommitteeNodes: bd.CommitteeNodes[:],
		AccessNodes:    bd.AccessNodes[:],
		Active:         bd.Active,
	}
}

func (bd *BootupData) BootupData() *registry.BootupData {
	return &registry.BootupData{
		ChainID:        *bd.ChainID.ChainID(),
		OwnerAddress:   *bd.OwnerAddress.Address(),
		Color:          *bd.Color.Color(),
		CommitteeNodes: bd.CommitteeNodes[:],
		AccessNodes:    bd.AccessNodes[:],
		Active:         bd.Active,
	}
}
