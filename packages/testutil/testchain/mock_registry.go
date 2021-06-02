package testchain

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/registry/committee_record"
)

type mockedCommitteeRegistry struct {
	validators []string
}

func NewMockedCommitteeRegistry(validators []string) *mockedCommitteeRegistry {
	return &mockedCommitteeRegistry{validators}
}

func (m *mockedCommitteeRegistry) GetCommitteeRecord(addr ledgerstate.Address) (*committee_record.CommitteeRecord, error) {
	return &committee_record.CommitteeRecord{
		Address: addr,
		Nodes:   m.validators,
	}, nil
}

func (m *mockedCommitteeRegistry) SaveCommitteeRecord(rec *committee_record.CommitteeRecord) error {
	panic("implement me")
}
