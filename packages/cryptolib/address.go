package cryptolib

import (
	"crypto/rand"
	"fmt"
	"io"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/bech32"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

const AddressSize = 32

type (
	Address    [AddressSize]byte
	AddressKey [AddressSize]byte
)

func NewEmptyAddress() *Address {
	return &Address{}
}

func NewRandomAddress() *Address {
	data := make([]byte, AddressSize)
	_, err := rand.Read(data)
	if err != nil {
		panic(fmt.Errorf("Failed to read random data, %s", err))
	}
	address, err := NewAddressFromBytes(data)
	if err != nil {
		panic(fmt.Errorf("Failed to create random address, %s", err))
	}
	return address
}

func newAddressFromArray(addr [AddressSize]byte) *Address {
	result := Address(addr)
	return &result
}

func NewAddressFromBytes(addr []byte) (*Address, error) {
	if len(addr) != AddressSize {
		return nil, fmt.Errorf("Array of size %v expected, size %v received", AddressSize, len(addr))
	}
	result := &Address{}
	copy(result[:], addr)
	return result, nil
}

func NewAddressFromString(addr string) (*Address, error) {
	addrBytes, err := DecodeHex(addr)
	if err != nil {
		return nil, fmt.Errorf("Error decoding hex: %w", err)
	}
	return NewAddressFromBytes(addrBytes)
}

// ParseBech32 decodes a bech32 encoded string.
func NewAddressFromBech32(s string) (iotago.NetworkPrefix, *Address, error) {
	hrp, addrData, err := bech32.Decode(s)
	if err != nil {
		return "", nil, fmt.Errorf("invalid bech32 encoding: %w", err)
	}

	if len(addrData) == 0 {
		return "", nil, fmt.Errorf("address data is empty")
	}

	address, err := NewAddressFromBytes(addrData)
	if err != nil {
		return "", nil, fmt.Errorf("failed to obtain address from byte array: %w", err)
	}

	return iotago.NetworkPrefix(hrp), address, nil
}

func NewAddressFromKey(key AddressKey) *Address {
	result := Address(key)
	return &result
}

// TODO: remove when not needed
func NewAddressFromIotago(addr iotago.Address) *Address {
	addrBytes, err := addr.Serialize(0, nil)
	if err != nil {
		panic(fmt.Sprintf("Failed to obtain byte array from iotago address: %s", err))
	}
	address, err := NewAddressFromBytes(addrBytes[1:])
	if err != nil {
		panic(fmt.Sprintf("Failed to obtain address from byte array: %s", err))
	}
	return address
}

// TODO: remove when not needed
func (a *Address) AsIotagoAddress() iotago.Address {
	result := iotago.Ed25519Address(a[:])
	return &result
}

func (a *Address) Equals(other *Address) bool {
	return *a == *other
}

func (a *Address) Bytes() []byte {
	return a[:]
}

func (a *Address) Bech32(hrp iotago.NetworkPrefix) string {
	s, err := bech32.Encode(string(hrp), a.Bytes())
	if err != nil {
		panic(err)
	}
	return s
}

func (a *Address) String() string {
	return EncodeHex(a.Bytes())
}

func (a *Address) Key() AddressKey {
	return AddressKey(*a)
}

func (a *Address) Clone() *Address {
	result := &Address{}
	copy(result[:], a[:])
	return result
}

func (a *Address) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	address := rr.ReadBytes()
	copy(a[:], address)
	return rr.Err
}

func (a *Address) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.WriteBytes(a[:])
	return ww.Err
}
