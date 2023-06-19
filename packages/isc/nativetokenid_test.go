package isc_test

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
)

func TestNativeTokenIDSerialization(t *testing.T) {
	data1 := make([]byte, len(iotago.NativeTokenID{}))
	rand.Read(data1)
	nativeTokenID1, err := isc.NativeTokenIDFromBytes(data1)
	require.NoError(t, err)
	hex1 := nativeTokenID1.ToHex()
	data2, err := iotago.DecodeHex(hex1)
	require.NoError(t, err)
	require.Equal(t, data1, data2)
	nativeTokenID2 := isc.MustNativeTokenIDFromBytes(data1)
	hex2 := nativeTokenID2.ToHex()
	data3, err := iotago.DecodeHex(hex2)
	require.NoError(t, err)
	require.Equal(t, data1, data3)
}
