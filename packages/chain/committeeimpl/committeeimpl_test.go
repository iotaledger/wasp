package committeeimpl

import (
	"testing"

	"github.com/iotaledger/wasp/packages/registry_pkg/committee_record"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/packages/peering"
	"github.com/iotaledger/wasp/packages/peering/udp"
	"github.com/iotaledger/wasp/packages/tcrypto"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/stretchr/testify/require"
	"go.dedis.ch/kyber/v3/pairing"
	"go.dedis.ch/kyber/v3/util/key"
)

func TestCommitteeBasic(t *testing.T) {
	suite := pairing.NewSuiteBn256()
	log := testlogger.NewLogger(t)
	defer log.Sync()
	netIDs := []string{"localhost:9017", "localhost:9018", "localhost:9019", "localhost:9020"}

	reg := newMockedRegistry(4, 3, netIDs)
	cfg0, err := peering.NewStaticPeerNetworkConfigProvider(netIDs[0], 9017, netIDs...)
	require.NoError(t, err)
	net0, err := udp.NewNetworkProvider(cfg0, key.NewKeyPair(suite), suite, log.Named("net0"))
	require.NoError(t, err)

	keyPair := ed25519.GenerateKeyPair()
	stateAddr := ledgerstate.NewED25519Address(keyPair.PublicKey)
	c, err := NewCommittee(stateAddr, net0, cfg0, reg, reg, log)
	require.NoError(t, err)
	require.True(t, c.Address().Equals(stateAddr))
	require.EqualValues(t, 4, c.Size())
	require.EqualValues(t, 3, c.Quorum())

	c.Close()
	require.False(t, c.IsReady())
}

type mockedRegistry struct {
	validators []string
	t, n       uint16
}

func newMockedRegistry(n, t uint16, validators []string) *mockedRegistry {
	return &mockedRegistry{validators, t, n}
}

func (m *mockedRegistry) SaveDKShare(dkShare *tcrypto.DKShare) error {
	panic("implement me")
}

func (m *mockedRegistry) LoadDKShare(sharedAddress ledgerstate.Address) (*tcrypto.DKShare, error) {
	var idx uint16
	return &tcrypto.DKShare{
		Address: sharedAddress,
		Index:   &idx,
		N:       4,
		T:       3,
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
