// Package coin implements cryptocurrency units and operations for the platform.
package coin

import (
	"encoding/json"
	"math"
	"math/big"
	"strconv"

	bcs "github.com/iotaledger/bcs-go"
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
)

// Value is the balance of a given coin
type Value uint64

func (v Value) Uint64() uint64 {
	return uint64(v)
}

func (v Value) BigInt() *big.Int {
	return new(big.Int).SetUint64(uint64(v))
}

func (v *Value) MarshalBCS(e *bcs.Encoder) error {
	e.WriteCompactUint64(uint64(*v))
	return nil
}

func (v *Value) UnmarshalBCS(d *bcs.Decoder) error {
	*v = Value(d.ReadCompactUint64())
	return nil
}

func (v Value) Bytes() []byte {
	return bcs.MustMarshal(&v)
}

func (v Value) String() string {
	return strconv.FormatUint(uint64(v), 10)
}

func ValueFromBytes(b []byte) (Value, error) {
	return bcs.Unmarshal[Value](b)
}

func ValueFromString(s string) (Value, error) {
	value, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return Value(0), err
	}

	return Value(value), nil
}

var BaseTokenType = MustTypeFromString(iotajsonrpc.IotaCoinType.String())

func IsBaseToken(t string) (bool, error) {
	return BaseTokenType.EqualsStr(t)
}

// Type is the representation of a Iota coin type, e.g. `0x000...0002::iota::IOTA`
// Two instances of Type are equal iif they represent the same coin type.
type Type = iotago.ObjectType

type TypeJSON = iotago.ObjectTypeJSON

func TypeFromString(s string) (Type, error) {
	return iotago.ObjectTypeFromString(s)
}

func MustTypeFromString(s string) Type {
	return iotago.MustTypeFromString(s)
}

func TypeFromBytes(b []byte) (Type, error) {
	return iotago.ObjectTypeFromBytes(b)
}

func CompareTypes(a, b Type) int {
	return iotago.CompareTypes(a, b)
}

var (
	Zero     = Value(0)
	MaxValue = Value(math.MaxUint64)
)

type CoinWithRef struct {
	Type  Type
	Value Value
	Ref   *iotago.ObjectRef
}

func (c CoinWithRef) String() string {
	b, err := json.MarshalIndent(&c, "", "\t")
	if err != nil {
		panic(err)
	}
	return string(b)
}

func (c CoinWithRef) Bytes() []byte {
	return bcs.MustMarshal(&c)
}
