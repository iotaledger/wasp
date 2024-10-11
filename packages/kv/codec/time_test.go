package codec

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestZeroTimeEncoding(t *testing.T) {
	z := time.Time{}
	require.True(t, z.IsZero())
	bin0 := Time.Encode(z)
	require.Equal(t, []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}, bin0)
	zback, err := Time.Decode(bin0)
	require.NoError(t, err)
	require.True(t, zback.IsZero())
	require.True(t, zback.Equal(z), "%v != %v", z, zback)
	require.True(t, zback.IsZero())
}
