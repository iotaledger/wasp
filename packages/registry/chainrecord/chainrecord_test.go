package chainrecord

import (
	"testing"

	"github.com/iotaledger/wasp/packages/coretypes/chainid"
	"github.com/stretchr/testify/require"
)

func TestChainRecord(t *testing.T) {
	chainID := chainid.RandomChainID()

	rec := ChainRecord{
		ChainID: chainID,
		Peers:   []string{"a", "b", "c"},
		Active:  false,
	}
	recBack, err := FromBytes(rec.Bytes())
	require.NoError(t, err)
	require.True(t, rec.ChainID.Equals(recBack.ChainID))
	require.EqualValues(t, rec.Active, recBack.Active)
	require.EqualValues(t, rec.Bytes(), recBack.Bytes())

	t.Logf("\n%s", rec.String())

	rec = ChainRecord{
		ChainID: chainID,
		Peers:   []string{"k", "l"},
		Active:  true,
	}
	require.True(t, rec.Active)
	recBack, err = FromBytes(rec.Bytes())
	require.NoError(t, err)
	require.True(t, rec.ChainID.Equals(recBack.ChainID))
	require.EqualValues(t, rec.Active, recBack.Active)
	require.EqualValues(t, rec.Bytes(), recBack.Bytes())
	t.Logf("\n%s", rec.String())
}
