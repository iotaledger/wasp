package codec_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/packages/kv/codec"
)

func TestOptional(t *testing.T) {
	b := codec.EncodeOptional[int64](nil)
	require.Equal(t, []byte{0x0}, b)

	vDec, err := codec.DecodeOptional[int64](b)
	require.NoError(t, err)
	require.Nil(t, vDec)

	v := int64(10)
	b = codec.EncodeOptional(&v)
	require.Equal(t, []byte{0x1, 10, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}, b)

	vDec, err = codec.DecodeOptional[int64](b)
	require.NoError(t, err)
	require.NotNil(t, vDec)
	require.Equal(t, int64(10), *vDec)
}

func TestSlice(t *testing.T) {
	b := codec.Encode([]int16{1, 2, 3})
	require.Equal(t, []byte{0x3, 0x1, 0x0, 0x2, 0x0, 0x3, 0x0}, b)

	v, err := codec.Decode[[]int16](b)
	require.NoError(t, err)
	require.Equal(t, []int16{1, 2, 3}, v)
}
