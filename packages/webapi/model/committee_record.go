package model

import (
	"github.com/iotaledger/wasp/packages/registry/committee_record"
)

type CommitteeRecord struct {
	Address Address  `swagger:"desc(Committee address (base58-encoded))"`
	Nodes   []string `swagger:"desc(List of committee nodes (network IDs))"`
}

func NewCommitteeRecord(bd *committee_record.CommitteeRecord) *CommitteeRecord {
	return &CommitteeRecord{
		Address: NewAddress(bd.Address),
		Nodes:   bd.Nodes,
	}
}

func (bd *CommitteeRecord) Record() *committee_record.CommitteeRecord {
	return &committee_record.CommitteeRecord{
		Address: bd.Address.Address(),
		Nodes:   bd.Nodes,
	}
}
