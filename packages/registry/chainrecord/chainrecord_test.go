package chainrecord

import (
	"testing"

	"github.com/iotaledger/wasp/packages/coretypes/chainid"
	"github.com/stretchr/testify/require"
)

func TestChainRecord(t *testing.T) {
	chainID := chainid.RandomChainID()

	rec := NewChainRecord(chainID, false)
	require.False(t, rec.Active)
	recBack, err := FromBytes(rec.Bytes())
	require.NoError(t, err)
	require.True(t, rec.ChainID.Equals(recBack.ChainID))
	require.EqualValues(t, rec.Active, recBack.Active)

	t.Logf("\n%s", rec)

	rec = NewChainRecord(chainID, true)
	require.True(t, rec.Active)
	recBack, err = FromBytes(rec.Bytes())
	require.NoError(t, err)
	require.True(t, rec.ChainID.Equals(recBack.ChainID))
	require.EqualValues(t, rec.Active, recBack.Active)
}
