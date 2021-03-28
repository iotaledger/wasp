package registry

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestChainRecord(t *testing.T) {
	keyPair := ed25519.GenerateKeyPair()
	stateAddr := ledgerstate.NewED25519Address(keyPair.PublicKey)
	chainID := coretypes.RandomChainID()

	rec := NewChainRecord(chainID, stateAddr)
	recBack, err := ChainRecordFromBytes(rec.Bytes())
	require.NoError(t, err)
	require.True(t, rec.ChainID.Equals(&recBack.ChainID))
	require.True(t, rec.StateAddressTmp.Equals(recBack.StateAddressTmp))
	require.EqualValues(t, rec.Active, recBack.Active)

	t.Logf("\n%s", rec)
}
