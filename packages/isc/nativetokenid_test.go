package isc_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/isc"
)

func TestNativeTokenIDSerialization(t *testing.T) {
	obj1 := isc.NativeTokenID("0x123")

	data1 := isc.NativeTokenIDToBytes(obj1)
	obj2, err := isc.NativeTokenIDFromBytes(data1)
	require.NoError(t, err)
	require.Equal(t, obj1, obj2)
	data2 := isc.NativeTokenIDToBytes(obj2)
	require.Equal(t, data1, data2)
}
