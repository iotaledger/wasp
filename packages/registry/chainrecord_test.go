// Copyright 2020 IOTA Stiftung
// SPDX-License-Identifier: Apache-2.0

package registry

import (
	"testing"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/stretchr/testify/require"
)

func TestChainRecord(t *testing.T) {
	chainID := isc.RandomChainID()

	rec := ChainRecord{
		ChainID: *chainID,
		Active:  false,
	}
	recBack, err := ChainRecordFromBytes(rec.Bytes())
	require.NoError(t, err)
	require.True(t, rec.ChainID.Equals(&recBack.ChainID))
	require.EqualValues(t, rec.Active, recBack.Active)
	require.EqualValues(t, rec.Bytes(), recBack.Bytes())

	t.Logf("\n%s", rec.String())

	rec = ChainRecord{
		ChainID: *chainID,
		Active:  true,
	}
	require.True(t, rec.Active)
	recBack, err = ChainRecordFromBytes(rec.Bytes())
	require.NoError(t, err)
	require.True(t, rec.ChainID.Equals(&recBack.ChainID))
	require.EqualValues(t, rec.Active, recBack.Active)
	require.EqualValues(t, rec.Bytes(), recBack.Bytes())
	t.Logf("\n%s", rec.String())
}
