package util_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/util"
)

func TestPrefixHex(t *testing.T) {
	require.Equal(t, "<nil>", util.PrefixHex(nil, 5))
	require.Equal(t, "0x010203", util.PrefixHex([]byte{1, 2, 3}, 5))
	require.Equal(t, "0x0102...", util.PrefixHex([]byte{1, 2, 3}, 2))
}
