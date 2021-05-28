package coretypes

import (
	"testing"

	"github.com/iotaledger/wasp/packages/registry_pkg/chain_record"
	"github.com/stretchr/testify/require"
)

func TestChainRecord(t *testing.T) {
	chainID := RandomChainID()

	rec := chain_record.NewChainRecord(chainID.AliasAddress, false, false)
	require.False(t, rec.Active)
	recBack, err := chain_record.ChainRecordFromBytes(rec.Bytes())
	require.NoError(t, err)
	require.True(t, rec.ChainIdAliasAddress.Equals(recBack.ChainIdAliasAddress))
	require.EqualValues(t, rec.Active, recBack.Active)

	t.Logf("\n%s", rec)

	rec = chain_record.NewChainRecord(chainID.AliasAddress, true, false)
	require.True(t, rec.Active)
	recBack, err = chain_record.ChainRecordFromBytes(rec.Bytes())
	require.NoError(t, err)
	require.True(t, rec.ChainIdAliasAddress.Equals(recBack.ChainIdAliasAddress))
	require.EqualValues(t, rec.Active, recBack.Active)
}
