package isc_test

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
)

func TestNativeTokenIDSerialization(t *testing.T) {
	obj1 := iotago.NativeTokenID{}
	rand.Read(obj1[:])

	data1 := isc.NativeTokenIDToBytes(obj1)
	obj2, err := isc.NativeTokenIDFromBytes(data1)
	require.NoError(t, err)
	require.Equal(t, obj1, obj2)
	data2 := isc.NativeTokenIDToBytes(obj2)
	require.Equal(t, data1, data2)

	hex1 := obj1.ToHex()
	data3, err := iotago.DecodeHex(hex1)
	require.NoError(t, err)
	require.Equal(t, data1, data3)
	obj3 := isc.MustNativeTokenIDFromBytes(data3)
	require.Equal(t, obj1, obj3)
}
