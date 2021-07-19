package testchain

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/registry"
)

type MockedCommitteeRegistry struct {
	validators []string
}

func NewMockedCommitteeRegistry(validators []string) *MockedCommitteeRegistry {
	return &MockedCommitteeRegistry{validators}
}

func (m *MockedCommitteeRegistry) GetCommitteeRecord(addr ledgerstate.Address) (*registry.CommitteeRecord, error) {
	return &registry.CommitteeRecord{
		Address: addr,
		Nodes:   m.validators,
	}, nil
}

func (m *MockedCommitteeRegistry) SaveCommitteeRecord(rec *registry.CommitteeRecord) error {
	panic("implement me")
}
