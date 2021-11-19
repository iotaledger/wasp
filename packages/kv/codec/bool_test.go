package codec

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func testBoolEncodeDecode(t *testing.T, b bool) {
	bin0 := EncodeBool(b)
	zback, err := DecodeBool(bin0)
	require.NoError(t, err)
	require.Equal(t, zback, b)
}

func TestBoolEncoding(t *testing.T) {
	testBoolEncodeDecode(t, true)
	testBoolEncodeDecode(t, false)
}
