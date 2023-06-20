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
		ed25519AddressByte1 := make([]byte, iotago.Ed25519AddressBytesLength+1)
		rand.Read(ed25519AddressByte1)
		ed25519AddressByte1[0] = byte(iotago.AddressEd25519)
		addr1, err := isc.AddressFromBytes(ed25519AddressByte1)
		require.NoError(t, err)
		ed25519AddressByte2 := isc.AddressToBytes(addr1)
		require.Equal(t, ed25519AddressByte1, ed25519AddressByte2)
	}
	{
		aliasAddressByte1 := make([]byte, iotago.AliasAddressBytesLength+1)
		rand.Read(aliasAddressByte1)
		aliasAddressByte1[0] = byte(iotago.AddressAlias)
		addr1, err := isc.AddressFromBytes(aliasAddressByte1)
		require.NoError(t, err)
		aliasAddressByte2 := isc.AddressToBytes(addr1)
		require.Equal(t, aliasAddressByte1, aliasAddressByte2)
	}
	{
		nftAddressByte1 := make([]byte, iotago.NFTAddressBytesLength+1)
		rand.Read(nftAddressByte1)
		nftAddressByte1[0] = byte(iotago.AddressNFT)
		addr1, err := isc.AddressFromBytes(nftAddressByte1)
		require.NoError(t, err)
		nftAddressByte2 := isc.AddressToBytes(addr1)
		require.Equal(t, nftAddressByte1, nftAddressByte2)
	}
}
