package sui

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"strings"

	"github.com/iotaledger/wasp/packages/util/rwutil"
)

const AddressLen = 32

type Address [AddressLen]uint8

func AddressFromArray(address [AddressLen]byte) *Address {
	result := Address(address)
	return &result
}

func AddressFromHex(str string) (*Address, error) {
	if strings.HasPrefix(str, "0x") || strings.HasPrefix(str, "0X") {
		str = str[2:]
	}
	if len(str)%2 != 0 {
		str = "0" + str
	}
	data, err := hex.DecodeString(str)
	if err != nil {
		return nil, err
	}
	if len(data) > AddressLen {
		return nil, errors.New("the len is invalid")
	}
	var address Address
	copy(address[AddressLen-len(data):], data[:])
	return &address, nil
}

func MustAddressFromHex(str string) *Address {
	addr, err := AddressFromHex(str)
	if err != nil {
		panic(err)
	}
	return addr
}

func RandomAddress() *Address {
	var a Address
	_, _ = rand.Read(a[:])
	return &a
}

func (a Address) Bytes() []byte { // NOTE: Bytes() is more conventional name for such method; I am not sure, if Data() method is necessary.
	return a[:]
}

func (a Address) Data() []byte {
	return a[:]
}

func (a Address) Length() int {
	return len(a)
}

func (a Address) String() string {
	return "0x" + hex.EncodeToString(a[:])
}

func (a Address) Equals(other Address) bool {
	return a == other
}

func (a Address) ToHex() string {
	return a.String()
}

func (a Address) ShortString() string {
	return "0x" + strings.TrimLeft(hex.EncodeToString(a[:]), "0")
}

func (a Address) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.String())
}

func (a *Address) UnmarshalJSON(data []byte) error {
	var str *string
	err := json.Unmarshal(data, &str)
	if err != nil {
		return err
	}
	if str == nil {
		return errors.New("nil address")
	}
	tmp, err := AddressFromHex(*str)
	if err == nil {
		*a = *tmp
	}
	return err
}

func (a *Address) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	address := rr.ReadBytes()
	copy(a[:], address)
	return rr.Err
}

func (a Address) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	ww.WriteBytes(a[:])
	return ww.Err
}

// FIXME may need to be pointer
func (a Address) MarshalBCS() ([]byte, error) {
	return a[:], nil
}
