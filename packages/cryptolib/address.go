package cryptolib

import (
	"fmt"

	"github.com/iotaledger/iota.go/v3/bech32"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

type AddressType byte

const (
	// AddressEd25519 denotes an Ed25519 address.
	AddressEd25519 AddressType = 0
	// AddressEd25519 denotes an Ed25519 address.
	AddressEd25519Legacy AddressType = 1
	// AddressAlias denotes an Alias address.
	AddressAlias AddressType = 8
	// AddressNFT denotes an NFT address.
	//AddressNFT AddressType = 16
	addressIsNil = 128
)

// Address describes a general address.
type Address interface {
	//serializer.SerializableWithSize
	//NonEphemeralObject
	fmt.Stringer

	// Type returns the type of the address.
	Type() AddressType

	Size() int
	Bytes() []byte

	// Bech32 encodes the address as a bech32 string.
	//Bech32(hrp NetworkPrefix) string

	// Equal checks whether other is equal to this Address.
	Equal(other Address) bool

	// Key returns a string which can be used to index the Address in a map.
	//Key() string

	// Clone clones the Address.
	//Clone() Address
}

func WriteAddress(a Address, ww *rwutil.Writer) {
	if a == nil {
		ww.WriteKind(addressIsNil)
		return
	}
	ww.WriteByte(byte(a.Type()))
	ww.WriteN(a.Bytes())
}

func SerializeAddress(a Address) []byte {
	ww := rwutil.NewBytesWriter()
	WriteAddress(a, ww)
	return ww.Bytes()
}

func ReadAddress(rr *rwutil.Reader) Address {
	addressType := rr.ReadByte()
	size := 0
	switch AddressType(addressType) {
	case AddressEd25519:
		size = Ed25519AddressBytesLength
	case AddressEd25519Legacy:
		size = Ed25519AddressLegacyBytesLength
	case AddressAlias:
		size = AliasAddressBytesLength
	default:
		size = 0
	}
	buf := make([]byte, size+1)
	buf[0] = addressType
	rr.ReadN(buf[1:])
	address, _ := DeserializeAddress(buf)
	return address
}

func DeserializeAddress(buf []byte) (Address, error) {
	if len(buf) == 0 {
		return nil, fmt.Errorf("empty address deserialization buffer")
	}
	addressType := buf[0]
	addressBuf := buf[1:]
	switch AddressType(addressType) {
	case AddressEd25519:
		return NewEd25519AddressFromBytes(addressBuf)
	case AddressEd25519Legacy:
		return NewEd25519AddressLegacyFromBytes(addressBuf)
	case AddressAlias:
		return NewAliasAddressFromBytes(addressBuf)
	case addressIsNil:
		return nil, nil
	default:
		return nil, nil
	}
}

// encodes the address as a bech32 string.
func AddressToBech32String(hrp parameters.NetworkPrefix, addr Address) string {
	bytes := SerializeAddress(addr)
	s, err := bech32.Encode(string(hrp), bytes)
	if err != nil {
		panic(err)
	}
	return s
}

// ParseBech32 decodes a bech32 encoded string.
func AddressFromBech32(s string) (parameters.NetworkPrefix, Address, error) {
	hrp, addrData, err := bech32.Decode(s)
	if err != nil {
		return "", nil, fmt.Errorf("invalid bech32 encoding: %w", err)
	}
	addr, err := DeserializeAddress(addrData)
	return parameters.NetworkPrefix(hrp), addr, err
}

// NonEphemeralObject is an object which can not be pruned by nodes as it
// makes up an integral part to execute the IOTA protocol. This kind of objects are associated
// with costs in terms of the resources they take up.
/*type NonEphemeralObject interface {
	// VBytes returns the cost this object has in terms of taking up
	// virtual and physical space within the data set needed to implement the IOTA protocol.
	// The override parameter acts as an escape hatch in case the cost needs to be adjusted
	// according to some external properties outside the NonEphemeralObject.
	VBytes(rentStruct *RentStructure, override VBytesFunc) VBytes
}*/
