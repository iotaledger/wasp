package cryptolib

import (
	"crypto/rand"
	"encoding/base64"
	"testing"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/stretchr/testify/require"
)

// $ sui keytool list
// ╭────────────────────────────────────────────────────────────────────────────────────────────╮
// │ ╭─────────────────┬──────────────────────────────────────────────────────────────────────╮ │
// │ │ alias           │  kind-chrysoberyl                                                    │ │
// │ │ suiAddress      │  0x1769589102f3749e02a24a1fcffea70e3fca86bf507cdb9d578a76a9c09e929a  │ │	<-- targetAddress
// │ │ publicBase64Key │  ALZfx6OH0xhJi8cXxoyyGhEi5A67byKdcHcaxdfskguH                        │ │	<-- publicKeyBase64
// │ │ keyScheme       │  ed25519                                                             │ │
// │ │ flag            │  0                                                                   │ │
// │ │ peerId          │  b65fc7a387d318498bc717c68cb21a1122e40ebb6f229d70771ac5d7ec920b87    │ │
// │ ╰─────────────────┴──────────────────────────────────────────────────────────────────────╯ │
// ╰────────────────────────────────────────────────────────────────────────────────────────────╯
func TestSuiPublicKeyToAddress(t *testing.T) {
	testSuiPublicKeyToAddress(t, "ALZfx6OH0xhJi8cXxoyyGhEi5A67byKdcHcaxdfskguH", "0x1769589102f3749e02a24a1fcffea70e3fca86bf507cdb9d578a76a9c09e929a")
	testSuiPublicKeyToAddress(t, "AOOU7NBSgm8KzQyh1wksLYu/b+IHG9u4U93QndtS7qIc", "0xcce6099545a2de67a326ad619e953f97e05cb4b6baaf2dcc74752265396922ef")
}

func testSuiPublicKeyToAddress(t *testing.T, publicKeyBase64, targetAddress string) {
	publicKeyWithType, err := base64.StdEncoding.DecodeString(publicKeyBase64)
	require.NoError(t, err)
	require.Equal(t, byte(0), publicKeyWithType[0])
	publicKey, err := PublicKeyFromBytes(publicKeyWithType[1:])
	require.NoError(t, err)
	require.Equal(t, targetAddress, EncodeHex(publicKey.AsAddress().Bytes()))
}

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
