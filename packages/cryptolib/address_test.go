package cryptolib

import (
	"crypto/rand"
	"testing"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/stretchr/testify/require"
)

func TestAddressSerialization(t *testing.T) {
	data := make([]byte, AddressSize)
	rand.Read(data)

	addr1, err := NewAddressFromBytes(data)
	require.NoError(t, err)

	data2 := addr1.Bytes()
	require.Equal(t, data, data2)
	addr2, err := NewAddressFromBytes(data2)
	require.NoError(t, err)
	require.True(t, addr1.Equals(addr2))
}

func TestAddressToIota(t *testing.T) {
	data := make([]byte, AddressSize)
	rand.Read(data)

	addr1, err := NewAddressFromBytes(data)
	require.NoError(t, err)

	addrIota := addr1.AsIotagoAddress()
	addr2 := NewAddressFromIotago(addrIota)
	require.True(t, addr1.Equals(addr2))
}

func TestAddressFromIota(t *testing.T) {
	data := make([]byte, iotago.Ed25519AddressBytesLength)
	rand.Read(data)

	addrIota1 := &iotago.Ed25519Address{}
	copy(addrIota1[:], data)

	addr := NewAddressFromIotago(addrIota1)
	addrIota2 := addr.AsIotagoAddress()

	require.True(t, addrIota1.Equal(addrIota2))
}
