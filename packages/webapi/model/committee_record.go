package model

import "github.com/iotaledger/wasp/packages/registry"

type CommitteeRecord struct {
	Address Address  `swagger:"desc(Committee address (base58-encoded))"`
	Nodes   []string `swagger:"desc(List of committee nodes (network IDs))"`
}

func NewCommitteeRecord(bd *registry.CommitteeRecord) *CommitteeRecord {
	return &CommitteeRecord{
		Address: NewAddress(bd.Address),
		Nodes:   bd.Nodes,
	}
}

func (bd *CommitteeRecord) Record() *registry.CommitteeRecord {
	return &registry.CommitteeRecord{
		Address: bd.Address.Address(),
		Nodes:   bd.Nodes,
	}
}
