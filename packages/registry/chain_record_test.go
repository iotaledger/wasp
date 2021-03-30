package registry

import (
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestChainRecord(t *testing.T) {
	chainID := coretypes.RandomChainID()

	rec := NewChainRecord(chainID)
	require.False(t, rec.Active)
	recBack, err := ChainRecordFromBytes(rec.Bytes())
	require.NoError(t, err)
	require.True(t, rec.ChainID.Equals(&recBack.ChainID))
	require.EqualValues(t, rec.Active, recBack.Active)

	t.Logf("\n%s", rec)

	rec = NewChainRecord(chainID, true)
	require.True(t, rec.Active)
	recBack, err = ChainRecordFromBytes(rec.Bytes())
	require.NoError(t, err)
	require.True(t, rec.ChainID.Equals(&recBack.ChainID))
	require.EqualValues(t, rec.Active, recBack.Active)
}
