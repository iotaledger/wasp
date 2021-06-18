package testchain

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/registry/committee_record"
)

type MockedCommitteeRegistry struct {
	validators []string
}

func NewMockedCommitteeRegistry(validators []string) *MockedCommitteeRegistry {
	return &MockedCommitteeRegistry{validators}
}

func (m *MockedCommitteeRegistry) GetCommitteeRecord(addr ledgerstate.Address) (*committee_record.CommitteeRecord, error) {
	return &committee_record.CommitteeRecord{
		Address: addr,
		Nodes:   m.validators,
	}, nil
}

func (m *MockedCommitteeRegistry) SaveCommitteeRecord(rec *committee_record.CommitteeRecord) error {
	panic("implement me")
}
