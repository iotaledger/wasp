package isc_test

import (
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/require"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
)

func TestAddressSerialization(t *testing.T) {
	{
		data := make([]byte, iotago.Ed25519AddressBytesLength+1)
		data[0] = byte(iotago.AddressEd25519)
		rand.Read(data[1:])

		addr1, err := isc.AddressFromBytes(data)
		require.NoError(t, err)
		require.IsType(t, &iotago.Ed25519Address{}, addr1)

		data2 := isc.AddressToBytes(addr1)
		require.Equal(t, data, data2)
		addr2, err := isc.AddressFromBytes(data2)
		require.NoError(t, err)
		require.Equal(t, addr1, addr2)
	}
	{
		data := make([]byte, iotago.AliasAddressBytesLength+1)
		data[0] = byte(iotago.AddressAlias)
		rand.Read(data[1:])

		addr1, err := isc.AddressFromBytes(data)
		require.NoError(t, err)
		require.IsType(t, &iotago.AliasAddress{}, addr1)

		data2 := isc.AddressToBytes(addr1)
		require.Equal(t, data, data2)
		addr2, err := isc.AddressFromBytes(data2)
		require.NoError(t, err)
		require.Equal(t, addr1, addr2)
	}
	{
		data := make([]byte, iotago.NFTAddressBytesLength+1)
		data[0] = byte(iotago.AddressNFT)
		rand.Read(data[1:])

		addr1, err := isc.AddressFromBytes(data)
		require.NoError(t, err)
		require.IsType(t, &iotago.NFTAddress{}, addr1)

		data2 := isc.AddressToBytes(addr1)
		require.Equal(t, data, data2)
		addr2, err := isc.AddressFromBytes(data2)
		require.NoError(t, err)
		require.Equal(t, addr1, addr2)
	}
}
