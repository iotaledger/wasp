package registry

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCommitteeRecord(t *testing.T) {
	keyPair := ed25519.GenerateKeyPair()
	addr := ledgerstate.NewED25519Address(keyPair.PublicKey)
	rec := NewCommitteeRecord(addr, "node:111", "node:333")
	recBack, err := CommitteeRecordFromBytes(rec.Bytes())
	require.NoError(t, err)
	require.True(t, rec.Address.Equals(recBack.Address))
	require.EqualValues(t, rec.CommitteeNodes, recBack.CommitteeNodes)

	t.Logf("%s", rec)
}
