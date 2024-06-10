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

func TestAddressToBech32(t *testing.T) {
	addr1 := NewRandomAddress()
	np1 := iotago.NetworkPrefix("abc")
	bech32 := addr1.Bech32(np1)
	np2, addr2, err := NewAddressFromBech32(bech32)
	require.NoError(t, err)
	require.Equal(t, np1, np2)
	require.True(t, addr1.Equals(addr2))
}

func TestAddressToString(t *testing.T) {
	addr1 := NewRandomAddress()
	str := addr1.String()
	addr2, err := NewAddressFromString(str)
	require.NoError(t, err)
	require.True(t, addr1.Equals(addr2))
}

func TestAddressToKey(t *testing.T) {
	addr1 := NewRandomAddress()
	key := addr1.Key()
	addr2 := NewAddressFromKey(key)
	require.True(t, addr1.Equals(addr2))
}

func TestAddressToIota(t *testing.T) {
	addr1 := NewRandomAddress()
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
