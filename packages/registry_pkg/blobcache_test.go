package registry_pkg

import (
	"testing"

	"github.com/iotaledger/wasp/packages/database/dbmanager"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/stretchr/testify/require"
)

func TestBlobPutGet(t *testing.T) {
	log := testlogger.NewLogger(t)
	dbmanager.CreateInstance(log, true)
	reg := NewRegistry(nil, log, nil)

	data := []byte("data-data-data-data-data-data-data-data-data")
	h := hashing.HashData(data)

	hback, err := reg.PutBlob(data)
	require.NoError(t, err)
	require.EqualValues(t, h, hback)

	back, ok, err := reg.GetBlob(h)
	require.NoError(t, err)
	require.True(t, ok)
	require.EqualValues(t, data, back)
}
