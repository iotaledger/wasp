package codec

import (
	"golang.org/x/xerrors"

	"github.com/iotaledger/wasp/packages/isc"
)

func DecodeAllowance(b []byte, def ...*isc.Allowance) (*isc.Allowance, error) {
	if b == nil {
		if len(def) == 0 {
			return nil, xerrors.New("cannot decode nil bytes")
		}
		return def[0], nil
	}
	return isc.AllowanceFromBytes(b)
}

func MustDecodeAllowance(b []byte, def ...*isc.Allowance) *isc.Allowance {
	ret, err := DecodeAllowance(b, def...)
	if err != nil {
		panic(err)
	}
	return ret
}

func EncodeAllowance(value *isc.Allowance) []byte {
	return value.Bytes()
}
