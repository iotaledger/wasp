package cryptolib

import (
	"crypto/rand"
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/require"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/v2/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/v2/packages/util/rwutil"
)

// $ iotago keytool list
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
func TestIotaPublicKeyToAddress(t *testing.T) {
	testIOTAPublicKeyToAddress(t, "ALZfx6OH0xhJi8cXxoyyGhEi5A67byKdcHcaxdfskguH", "0x1769589102f3749e02a24a1fcffea70e3fca86bf507cdb9d578a76a9c09e929a")
	testIOTAPublicKeyToAddress(t, "AOOU7NBSgm8KzQyh1wksLYu/b+IHG9u4U93QndtS7qIc", "0xcce6099545a2de67a326ad619e953f97e05cb4b6baaf2dcc74752265396922ef")
}

func testIOTAPublicKeyToAddress(t *testing.T, publicKeyBase64, targetAddress string) {
	publicKeyWithType, err := base64.StdEncoding.DecodeString(publicKeyBase64)
	require.NoError(t, err)
	require.Equal(t, byte(0), publicKeyWithType[0])
	publicKey, err := PublicKeyFromBytes(publicKeyWithType)
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

func TestAddressToString(t *testing.T) {
	addr1 := NewRandomAddress()
	str := addr1.String()
	addr2, err := NewAddressFromHexString(str)
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
	addrIota := addr1.AsIotaAddress()
	addr2 := NewAddressFromIota(addrIota)
	require.True(t, addr1.Equals(addr2))
}

func TestAddressFromIota(t *testing.T) {
	addrIota1 := iotago.Address{}
	rand.Read(addrIota1[:])

	addr := NewAddressFromIota(&addrIota1)
	addrIota2 := addr.AsIotaAddress()

	require.True(t, addrIota1.Equals(*addrIota2))
}

func TestAddressBCSCodec(t *testing.T) {
	addr := NewRandomAddress()
	encBCS := bcs.MustMarshal(&addr)
	encRwutil := rwutil.NewBytesWriter().Write(addr).Bytes()

	require.Equal(t, encBCS, encRwutil, addr)
	require.Equal(t, len(encBCS), len(addr), encBCS)
}
