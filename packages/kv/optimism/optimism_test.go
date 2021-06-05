package optimism

import (
	"testing"

	"github.com/iotaledger/wasp/packages/coretypes/coreutil"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/stretchr/testify/require"
)

func TestOptimismBasic(t *testing.T) {
	d := dict.New()
	d.Set("a", []byte("b"))
	d.Set("c", []byte("d"))
	glb := coreutil.NewGlobalSync()
	base := glb.GetSolidIndexBaseline()
	require.False(t, base.IsValid())
	glb.SetSolidIndex(2)
	require.False(t, base.IsValid())
	base.SetBaseline()
	require.True(t, base.IsValid())

	r := NewOptimisticKVStoreReader(d, base)

	b, err := r.Get("a")
	require.NoError(t, err)
	require.EqualValues(t, "b", string(b))

	glb.SetSolidIndex(3)
	require.False(t, base.IsValid())
	_, err = r.Get("a")
	require.Error(t, err)
	require.EqualValues(t, err, ErrStateHasBeenInvalidated)

	r.SetBaseline()
	b, err = r.Get("a")
	require.NoError(t, err)
	require.EqualValues(t, "b", string(b))

	glb.InvalidateSolidIndex()
	_, err = r.Get("a")
	require.Error(t, err)
	require.EqualValues(t, err, ErrStateHasBeenInvalidated)
}
