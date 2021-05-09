package mock_chain

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/registry_pkg/committee_record"
	"github.com/iotaledger/wasp/packages/tcrypto"
)

type mockedRegistry struct {
	validators     []string
	t, n, ownIndex uint16
}

func NewMockedRegistry(n, t, ownIndex uint16, validators []string) *mockedRegistry {
	return &mockedRegistry{validators, t, n, ownIndex}
}

func (m *mockedRegistry) SaveDKShare(dkShare *tcrypto.DKShare) error {
	panic("implement me")
}

func (m *mockedRegistry) LoadDKShare(sharedAddress ledgerstate.Address) (*tcrypto.DKShare, error) {
	return &tcrypto.DKShare{
		Address: sharedAddress,
		Index:   &m.ownIndex,
		N:       m.n,
		T:       m.t,
	}, nil
}

func (m *mockedRegistry) GetCommitteeRecord(addr ledgerstate.Address) (*committee_record.CommitteeRecord, error) {
	return &committee_record.CommitteeRecord{
		Address: addr,
		Nodes:   m.validators,
	}, nil
}

func (m *mockedRegistry) SaveCommitteeRecord(rec *committee_record.CommitteeRecord) error {
	panic("implement me")
}
