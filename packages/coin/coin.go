package coin

import (
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"

	"github.com/samber/lo"

	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/clients/iota-go/iotajsonrpc"
	"github.com/iotaledger/wasp/packages/util/bcs"
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
	_ = e.WriteCompactUint(uint64(*v))
	return e.Err()
}

func (v *Value) UnmarshalBCS(d *bcs.Decoder) error {
	*v = Value(d.ReadCompactUint())
	return d.Err()
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

// TODO: maybe it is not ok to consider this constant?
var BaseTokenType = MustTypeFromString(iotajsonrpc.IotaCoinType.String())

// Type is the representation of a Iota coin type, e.g. `0x000...0002::iota::IOTA`
// Two instances of Type are equal iif they represent the same coin type.
type Type struct { // struct to enforce using the constructor functions
	s string
}

func TypeFromString(s string) (Type, error) {
	rt, err := iotago.NewResourceType(s)
	if err != nil {
		return Type{}, fmt.Errorf("invalid Type %q: %w", s, err)
	}
	return Type{s: rt.String()}, nil
}

func MustTypeFromString(s string) Type {
	t, err := TypeFromString(s)
	if err != nil {
		panic(err)
	}
	return t
}

func (t *Type) MarshalBCS(e *bcs.Encoder) error {
	rt := t.ResourceType()
	_ = e.Encode(rt)
	return e.Err()
}

func (t *Type) UnmarshalBCS(d *bcs.Decoder) error {
	var rt iotago.ResourceType
	_ = d.Decode(&rt)
	if d.Err() != nil {
		return d.Err()
	}
	*t = Type{s: rt.String()}
	return nil
}

// MatchesStringType returns true if the given string represents the same coin
// type, even if abbreviated (e.g. ""0x2::iota::IOTA"")
func (t Type) MatchesStringType(s string) bool {
	rt, err := TypeFromString(s)
	if err != nil {
		return false
	}
	return rt.String() == t.s
}

func (t Type) String() string {
	return t.s
}

func (t Type) AsRPCCoinType() iotajsonrpc.CoinType {
	return iotajsonrpc.CoinType(t.String())
}

func (t Type) ShortString() string {
	return t.ResourceType().ShortString()
}

func (t Type) ResourceType() *iotago.ResourceType {
	return lo.Must(iotago.NewResourceType(t.s))
}

func (t Type) TypeTag() iotago.TypeTag {
	coinTypeTag, err := iotago.TypeTagFromString(t.String())
	if err != nil {
		panic(err)
	}
	return *coinTypeTag
}

func (t Type) Bytes() []byte {
	return bcs.MustMarshal(&t)
}

func TypeFromBytes(b []byte) (Type, error) {
	var r Type
	r, err := bcs.Unmarshal[Type](b)
	return r, err
}

func (t Type) ToIotaJSONRPC() iotajsonrpc.CoinType {
	return iotajsonrpc.CoinType(t.String())
}

func CompareTypes(a, b Type) int {
	return strings.Compare(a.s, b.s)
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

func (c CoinWithRef) Bytes() []byte {
	return bcs.MustMarshal(&c)
}
