package codec_test

import (
	"testing"

	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/stretchr/testify/require"
)

func TestSliceToFrom(t *testing.T) {
	c := codec.NewCodecFromBCS[int16]()

	b := codec.SliceToArray(c, []int16{1, 2, 3})
	require.Equal(t, []byte{0x3, 0x1, 0x0, 0x2, 0x0, 0x3, 0x0}, b)

	v, err := codec.SliceFromArray(c, b)
	require.NoError(t, err)
	require.Equal(t, []int16{1, 2, 3}, v)
}
