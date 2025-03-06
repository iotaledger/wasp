package bcs_test

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/util/bcs"
)

func TestUint128(t *testing.T) {
	testUint128Codec(t, "10", true, []byte{0xa, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0})
	testUint128Codec(t, "1770887431076116955186", true, []byte{0x32, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x60, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0})
	testUint128Codec(t, "999999999999999999999999999999999999999999999999999", false)
}

func testUint128Codec(t *testing.T, v string, expectSuccess bool, expectedBytes ...[]byte) {
	var bi big.Int
	_, ok := bi.SetString(v, 10)
	require.True(t, ok)

	if expectSuccess {
		bcs.TestCodecAndBytes(t, bi, expectedBytes[0])
		bcs.TestCodecAndBytes(t, &bi, expectedBytes[0])
	} else {
		bcs.TestEncodeErr(t, bi)
		bcs.TestEncodeErr(t, &bi)
	}
}
