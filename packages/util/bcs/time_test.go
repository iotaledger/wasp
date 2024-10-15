package bcs_test

import (
	"testing"
	"time"

	"github.com/iotaledger/wasp/packages/util/bcs"
	"github.com/stretchr/testify/require"
)

func TestTimeCodec(t *testing.T) {
	bcs.TestCodecAndBytes(t, time.Unix(12345, 6789), []byte{0x85, 0x14, 0x57, 0x4b, 0x3a, 0xb, 0x0, 0x0})

	encoded := bcs.MustMarshal(new(time.Time))
	require.Equal(t, []byte{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}, encoded)
	decoded := bcs.MustUnmarshal[time.Time](encoded)
	require.True(t, decoded.IsZero())
}
