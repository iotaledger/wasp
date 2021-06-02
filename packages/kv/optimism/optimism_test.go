package optimism

import (
	"testing"

	"github.com/iotaledger/wasp/packages/coretypes/coreutil"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"
)

func TestOptimismBasic(t *testing.T) {
	d := dict.New()
	d.Set("a", []byte("b"))
	d.Set("c", []byte("d"))
	at := atomic.NewUint32(0)
	base := coreutil.NewStateIndexBaseline(at)
	r := NewOptimisticKVStoreReader(d, base)
	require.True(t, base.IsValid())

	b, err := r.Get("a")
	require.NoError(t, err)
	require.EqualValues(t, "b", string(b))

	at.Store(2)
	require.False(t, base.IsValid())
	_, err = r.Get("a")
	require.Error(t, err)
	require.EqualValues(t, err, ErrStateHasBeenInvalidated)
}
