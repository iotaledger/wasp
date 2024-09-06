package coin

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/util/bcs"
	"github.com/iotaledger/wasp/sui-go/sui"
	"github.com/iotaledger/wasp/sui-go/suijsonrpc"
)

// Value is the balance of a given coin
type Value uint64

func (v Value) Bytes() []byte {
	return bcs.MustMarshal(&v)
}

func ValueFromBytes(b []byte) (Value, error) {
	return bcs.Unmarshal[Value](b)
}

// TODO: maybe it is not ok to consider this constant?
const BaseTokenType = Type(suijsonrpc.SuiCoinType)

// Type is the string representation of a Sui coin type, e.g. `0x2::sui::SUI`
type Type string

func (t Type) MarshalBCS(e *bcs.Encoder) error {
	rt, err := sui.NewResourceType(string(t))
	if err != nil {
		return fmt.Errorf("invalid Type %q: %w", t, err)
	}
	if rt.SubType != nil {
		panic("cointype with subtype is unsupported")
	}

	return e.Encode(rt)
}

func (t *Type) UnmarshalBCS(e *bcs.Decoder) error {
	rt := sui.ResourceType{}
	if err := e.Decode(&rt); err != nil {
		return err
	}
	if rt.SubType != nil {
		panic("cointype with subtype is unsupported")
	}
	*t = Type(rt.ShortString())

	return nil
}

func (t Type) String() string {
	return string(t)
}

func (t Type) Bytes() []byte {
	return bcs.MustMarshal(&t)
}

func TypeFromBytes(b []byte) (Type, error) {
	return bcs.Unmarshal[Type](b)
}
