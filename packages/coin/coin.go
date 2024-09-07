package coin

import (
	"errors"
	"fmt"
	"io"
	"math/big"

	"github.com/iotaledger/wasp/packages/util/rwutil"
	"github.com/iotaledger/wasp/sui-go/sui"
	"github.com/iotaledger/wasp/sui-go/suijsonrpc"
)

// Value is the balance of a given coin
type Value uint64

func (v Value) Uint64() uint64 {
	return uint64(v)
}

func (v Value) Write(w io.Writer) error {
	ww := rwutil.NewWriter(w)
	// serialized as bigint to save space for smaller values
	ww.WriteBigUint(new(big.Int).SetUint64(uint64(v)))
	return ww.Err
}

func (v *Value) Read(r io.Reader) error {
	rr := rwutil.NewReader(r)
	b := rr.ReadBigUint()
	if rr.Err != nil {
		return rr.Err
	}
	if !b.IsUint64() {
		return errors.New("cannot read Value: must be uint64")
	}
	*v = Value(b.Uint64())
	return nil
}

func (v Value) Bytes() []byte {
	return rwutil.WriteToBytes(v)
}

func ValueFromBytes(b []byte) (Value, error) {
	var r Value
	_, err := rwutil.ReadFromBytes(b, &r)
	return r, err
}

// TODO: maybe it is not ok to consider this constant?
const BaseTokenType = Type(suijsonrpc.SuiCoinType)

// Type is the string representation of a Sui coin type, e.g. `0x2::sui::SUI`
type Type string

func (t Type) Write(w io.Writer) error {
	rt, err := sui.NewResourceType(string(t))
	if err != nil {
		return fmt.Errorf("invalid Type %q: %w", t, err)
	}
	if rt.SubType1 != nil {
		panic("cointype with subtype is unsupported")
	}
	ww := rwutil.NewWriter(w)
	ww.WriteN(rt.Address[:])
	ww.WriteString(rt.Module)
	ww.WriteString(rt.ObjectName)
	return ww.Err
}

func (t *Type) Read(r io.Reader) error {
	rt := sui.ResourceType{
		Address: &sui.Address{},
	}
	rr := rwutil.NewReader(r)
	rr.ReadN(rt.Address[:])
	rt.Module = rr.ReadString()
	rt.ObjectName = rr.ReadString()
	*t = Type(rt.ShortString())
	return rr.Err
}

func (t Type) String() string {
	return string(t)
}

func (t Type) TypeTag() sui.TypeTag {
	coinTypeTag, err := sui.TypeTagFromString(t.String())
	if err != nil {
		panic(err)
	}
	return *coinTypeTag
}

func (t Type) Bytes() []byte {
	return rwutil.WriteToBytes(t)
}

func TypeFromBytes(b []byte) (Type, error) {
	var r Type
	_, err := rwutil.ReadFromBytes(b, &r)
	return r, err
}
