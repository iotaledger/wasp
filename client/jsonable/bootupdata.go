package jsonable

import (
	"github.com/iotaledger/wasp/packages/registry"
)

type BootupData struct {
	Address        *Address
	OwnerAddress   *Address
	Color          *Color
	CommitteeNodes []string
	AccessNodes    []string
	Active         bool
}

func NewBootupData(bd *registry.BootupData) *BootupData {
	return &BootupData{
		Address:        NewAddress(&bd.Address),
		OwnerAddress:   NewAddress(&bd.OwnerAddress),
		Color:          NewColor(&bd.Color),
		CommitteeNodes: bd.CommitteeNodes[:],
		AccessNodes:    bd.AccessNodes[:],
		Active:         bd.Active,
	}
}

func (bd *BootupData) BootupData() *registry.BootupData {
	return &registry.BootupData{
		Address:        *bd.Address.Address(),
		OwnerAddress:   *bd.OwnerAddress.Address(),
		Color:          *bd.Color.Color(),
		CommitteeNodes: bd.CommitteeNodes[:],
		AccessNodes:    bd.AccessNodes[:],
		Active:         bd.Active,
	}
}
