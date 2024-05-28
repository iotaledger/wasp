package cryptolib

import (
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

func newAddressFromArray(addr [AddressSize]byte) *Address {
	result := Address(addr)
	return &result
}

func NewAddressFromBytes(addr []byte) *Address {
	// TODO: check slice length?
	result := &Address{}
	copy(result[:], addr)
	return result
}

func NewAddressFromString(addr string) (*Address, error) {
	addrBytes, err := DecodeHex(addr)
	if err != nil {
		return nil, err
	}
	return NewAddressFromBytes(addrBytes), nil
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

	return iotago.NetworkPrefix(hrp), NewAddressFromBytes(addrData), nil
}

// TODO: remove when not needed
func NewAddressFromIotago(addr iotago.Address) *Address {
	addrBytes, _ := addr.Serialize(0, nil)
	return NewAddressFromBytes(addrBytes)
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
