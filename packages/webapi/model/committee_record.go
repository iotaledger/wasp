package model

import (
	"github.com/iotaledger/wasp/packages/registry"
)

type CommitteeRecord struct {
	Address        Address  `swagger:"desc(Committee address (base58-encoded))"`
	CommitteeNodes []string `swagger:"desc(List of committee nodes (network IDs))"`
}

func NewCommitteeRecord(bd *registry.CommitteeRecord) *CommitteeRecord {
	return &CommitteeRecord{
		Address:        NewAddress(bd.Address),
		CommitteeNodes: bd.CommitteeNodes,
	}
}

func (bd *CommitteeRecord) Record() *registry.CommitteeRecord {
	return &registry.CommitteeRecord{
		Address:        bd.Address.Address(),
		CommitteeNodes: bd.CommitteeNodes,
	}
}
