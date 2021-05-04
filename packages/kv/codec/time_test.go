package codec

import (
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestZeroTimeEncoding(t *testing.T) {
	require.EqualValues(t, zeroUnixNano, -6795364578871345152)
	z := time.Time{}
	require.EqualValues(t, z.UnixNano(), zeroUnixNano)
	require.True(t, z.IsZero())
	bin0 := EncodeTime(z)
	zback, ok, err := DecodeTime(bin0)
	require.NoError(t, err)
	require.True(t, ok)
	require.True(t, zback.IsZero())
	require.True(t, zback.Equal(z))
	require.EqualValues(t, zback.UnixNano(), zeroUnixNano)
}
