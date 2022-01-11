package registry

import (
	"testing"

	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/stretchr/testify/require"
)

func TestCommitteeRecord(t *testing.T) {
	keyPair := cryptolib.NewKeyPair()
	addr := ledgerstate.NewED25519Address(keyPair.PublicKey)
	rec := NewCommitteeRecord(addr, "node:111", "node:333")
	recBack, err := CommitteeRecordFromBytes(rec.Bytes())
	require.NoError(t, err)
	require.True(t, rec.Address.Equals(recBack.Address))
	require.EqualValues(t, rec.Nodes, recBack.Nodes)

	t.Logf("%s", rec)
}
