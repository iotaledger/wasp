package optimism

import (
	"testing"

	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/stretchr/testify/require"
)

func TestOptimismBasic(t *testing.T) {
	d := dict.New()
	d.Set("a", []byte("b"))
	d.Set("c", []byte("d"))
	glb := coreutil.NewChainStateSync()
	base := glb.GetSolidIndexBaseline()
	require.False(t, base.IsValid())
	glb.SetSolidIndex(2)
	require.False(t, base.IsValid())
	base.Set()
	require.True(t, base.IsValid())

	r := NewOptimisticKVStoreReader(d, base)

	b, err := r.Get("a")
	require.NoError(t, err)
	require.EqualValues(t, "b", string(b))

	glb.SetSolidIndex(3)
	require.False(t, base.IsValid())
	_, err = r.Get("a")
	require.Error(t, err)
	require.ErrorIs(t, err, coreutil.ErrorStateInvalidated)

	r.SetBaseline()
	b, err = r.Get("a")
	require.NoError(t, err)
	require.EqualValues(t, "b", string(b))

	glb.InvalidateSolidIndex()
	_, err = r.Get("a")
	require.Error(t, err)
	require.ErrorIs(t, err, coreutil.ErrorStateInvalidated)
}

// returns the number of times the read operation was executed
func tryReadFromDB(storeReader *OptimisticKVStoreReader) (int, error) {
	readCounter := 0
	err := RetryOnStateInvalidated(func() error {
		var err error
		readCounter++
		_, err = storeReader.Get("foo")
		return err
	})
	return readCounter, err
}

func TestRetryOnStateInvalidation(t *testing.T) {
	d := dict.New()
	d.Set("foo", []byte("bar"))
	glb := coreutil.NewChainStateSync()
	base := glb.GetSolidIndexBaseline()
	r := NewOptimisticKVStoreReader(d, base)

	glb.SetSolidIndex(0)
	base.Set()
	require.True(t, base.IsValid())

	nReads, err := tryReadFromDB(r)
	require.NoError(t, err)
	require.EqualValues(t, 1, nReads)

	glb.SetSolidIndex(3)
	require.False(t, base.IsValid())

	nReads, err = tryReadFromDB(r)
	require.ErrorIs(t, err, coreutil.ErrorStateInvalidated)
	require.Greater(t, nReads, 1)
}
