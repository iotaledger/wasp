package chainrecord

import (
	"testing"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/stretchr/testify/require"
)

func TestChainRecord(t *testing.T) {
	randombytes := hashing.RandomHash(nil)
	AliasAddress := ledgerstate.NewAliasAddress(randombytes[:])

	rec := NewChainRecord(AliasAddress, false)
	require.False(t, rec.Active)
	recBack, err := ChainRecordFromBytes(rec.Bytes())
	require.NoError(t, err)
	require.True(t, rec.ChainAddr.Equals(recBack.ChainAddr))
	require.EqualValues(t, rec.Active, recBack.Active)

	t.Logf("\n%s", rec)

	rec = NewChainRecord(AliasAddress, true)
	require.True(t, rec.Active)
	recBack, err = ChainRecordFromBytes(rec.Bytes())
	require.NoError(t, err)
	require.True(t, rec.ChainAddr.Equals(recBack.ChainAddr))
	require.EqualValues(t, rec.Active, recBack.Active)
}
