package optimism

import (
	"testing"
	"time"

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
	require.Equal(t, err, coreutil.ErrorStateInvalidated)

	r.SetBaseline()
	require.True(t, base.IsValid())
	b, err = r.Get("a")
	require.NoError(t, err)
	require.EqualValues(t, "b", string(b))

	glb.InvalidateSolidIndex()
	_, err = r.Get("a")
	require.Error(t, err)
	require.Equal(t, err, coreutil.ErrorStateInvalidated)
}

// returns the number of times the read operation was executed
func tryReadFromDB(storeReader *OptimisticKVStoreReader, timeouts ...time.Duration) ([]byte, int, error) {
	readCounter := 0
	var ret []byte
	err := RetryOnStateInvalidated(func() error {
		var err error
		readCounter++
		ret, err = storeReader.Get("foo")
		return err
	}, timeouts...)
	return ret, readCounter, err
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

	_, nReads, err := tryReadFromDB(r)
	require.NoError(t, err)
	require.EqualValues(t, 1, nReads)

	glb.SetSolidIndex(3)
	require.False(t, base.IsValid())

	_, nReads, err = tryReadFromDB(r)
	require.NotEqual(t, err, coreutil.ErrorStateInvalidated)
	require.Greater(t, nReads, 1)

	// make sure it stops retrying and returns once a new baseline is set
	go func() {
		ret, nReads, err := tryReadFromDB(r, 5*time.Second, 10*time.Millisecond)
		require.NoError(t, err)
		require.Greater(t, nReads, 1)
		require.Equal(t, ret, []byte("bar"))
	}()

	time.Sleep(50 * time.Millisecond)
	r.SetBaseline() // sets the baseline while the go-routine above is running, so
}
