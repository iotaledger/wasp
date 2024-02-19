package coreutil

import (
	"fmt"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

func Optional[T any](v ...T) *T {
	if len(v) > 0 {
		return &v[0]
	}
	return nil
}

func FromOptional[T any](opt *T, def T) T {
	if opt == nil {
		return def
	}
	return *opt
}

// DictCodec is the interface for any type that can be converted to/from dict.Dict
type DictCodec[T any] interface {
	Encode(T) dict.Dict
	Decode(dict.Dict) (T, error)
}

// RawDictCodec is a DictCodec that performs no conversion
type RawDictCodec struct{}

func (RawDictCodec) Decode(d dict.Dict) (dict.Dict, error) {
	return d, nil
}

func (RawDictCodec) Encode(d dict.Dict) dict.Dict {
	return d
}

// Field is a DictCodec that converts a single value into a single dict key
type Field[T any] struct {
	Key   kv.Key
	Codec codec.Codec[T]
}

func (f Field[T]) Encode(v T) dict.Dict {
	b := f.Codec.Encode(v)
	if b == nil {
		return dict.Dict{}
	}
	return dict.Dict{f.Key: b}
}

func (f Field[T]) Decode(d dict.Dict) (T, error) {
	return f.Codec.Decode(d[f.Key])
}

func FieldWithCodec[T any](key kv.Key, codec codec.Codec[T]) Field[T] {
	return Field[T]{Key: key, Codec: codec}
}

type OptionalCodec[T any] struct {
	codec.Codec[T]
}

func (c *OptionalCodec[T]) Decode(b []byte, def ...*T) (r *T, err error) {
	if b == nil {
		if len(def) != 0 {
			err = fmt.Errorf("%T: unexpected default value", r)
			return
		}
		return nil, nil
	}
	v, err := c.Codec.Decode(b)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (c *OptionalCodec[T]) Encode(v *T) []byte {
	if v == nil {
		return nil
	}
	return c.Codec.Encode(*v)
}

// FieldWithCodecOptional returns a Field that accepts an optional value
func FieldWithCodecOptional[T any](key kv.Key, c codec.Codec[T]) Field[*T] {
	return Field[*T]{Key: key, Codec: &OptionalCodec[T]{Codec: c}}
}

// EP0 is a utility type for entry points that receive 0 parameters
type EP0[S isc.SandboxBase] struct{ EntryPointInfo[S] }

func (e EP0[S]) Message() isc.Message { return e.EntryPointInfo.Message(nil) }

func NewEP0(contract *ContractInfo, name string) EP0[isc.Sandbox] {
	return EP0[isc.Sandbox]{EntryPointInfo: contract.Func(name)}
}

func NewViewEP0(contract *ContractInfo, name string) EP0[isc.SandboxView] {
	return EP0[isc.SandboxView]{EntryPointInfo: contract.ViewFunc(name)}
}

// EP1 is a utility type for entry points that receive 1 parameter
type EP1[S isc.SandboxBase, T any, I DictCodec[T]] struct {
	EntryPointInfo[S]
	Input I
}

func (e EP1[S, T, I]) Message(p1 T) isc.Message {
	return e.EntryPointInfo.Message(e.Input.Encode(p1))
}

func (e EP1[S, T, I]) WithHandler(f func(ctx S, p T) dict.Dict) *EntryPointHandler[S] {
	return e.EntryPointInfo.WithHandler(func(ctx S) dict.Dict {
		p, err := e.Input.Decode(ctx.Params().Dict)
		ctx.RequireNoError(err)
		return f(ctx, p)
	})
}

func NewEP1[T any, I DictCodec[T]](contract *ContractInfo, name string, in I) EP1[isc.Sandbox, T, I] {
	return EP1[isc.Sandbox, T, I]{
		EntryPointInfo: contract.Func(name),
		Input:          in,
	}
}

func NewViewEP1[T any, I DictCodec[T]](contract *ContractInfo, name string, in I) EP1[isc.SandboxView, T, I] {
	return EP1[isc.SandboxView, T, I]{
		EntryPointInfo: contract.ViewFunc(name),
		Input:          in,
	}
}

// EP2 is a utility type for entry points that receive 2 parameters
type EP2[S isc.SandboxBase, T1, T2 any, I1 DictCodec[T1], I2 DictCodec[T2]] struct {
	EntryPointInfo[S]
	Input1 I1
	Input2 I2
}

func NewEP2[T1 any, T2 any, I1 DictCodec[T1], I2 DictCodec[T2]](
	contract *ContractInfo, name string,
	in1 I1,
	in2 I2,
) EP2[isc.Sandbox, T1, T2, I1, I2] {
	return EP2[isc.Sandbox, T1, T2, I1, I2]{
		EntryPointInfo: contract.Func(name),
		Input1:         in1,
		Input2:         in2,
	}
}

func NewViewEP2[T1 any, T2 any, I1 DictCodec[T1], I2 DictCodec[T2]](
	contract *ContractInfo, name string,
	in1 I1,
	in2 I2,
) EP2[isc.SandboxView, T1, T2, I1, I2] {
	return EP2[isc.SandboxView, T1, T2, I1, I2]{
		EntryPointInfo: contract.ViewFunc(name),
		Input1:         in1,
		Input2:         in2,
	}
}

func (e EP2[S, T1, T2, I1, I2]) WithHandler(f func(ctx S, p1 T1, p2 T2) dict.Dict) *EntryPointHandler[S] {
	return e.EntryPointInfo.WithHandler(func(ctx S) dict.Dict {
		params := ctx.Params()
		p1, err := e.Input1.Decode(params.Dict)
		ctx.RequireNoError(err)
		p2, err := e.Input2.Decode(params.Dict)
		ctx.RequireNoError(err)
		return f(ctx, p1, p2)
	})
}

func (e EP2[S, T1, T2, I1, I2]) Message(p1 T1, p2 T2) isc.Message {
	p := e.Input1.Encode(p1)
	p.Extend(e.Input2.Encode(p2))
	return e.EntryPointInfo.Message(p)
}

// EP01 is a utility type for entry points that receive 0 parameters and return 1 value
type EP01[S isc.SandboxBase, R any, O DictCodec[R]] struct {
	EP0[S]
	Output O
}

func NewViewEP01[R any, O DictCodec[R]](
	contract *ContractInfo, name string,
	out O,
) EP01[isc.SandboxView, R, O] {
	return EP01[isc.SandboxView, R, O]{
		EP0:    NewViewEP0(contract, name),
		Output: out,
	}
}

func (e EP01[S, R, O]) WithHandler(f func(ctx S) R) *EntryPointHandler[S] {
	return e.EntryPointInfo.WithHandler(func(ctx S) dict.Dict {
		r := f(ctx)
		return e.Output.Encode(r)
	})
}

// EP02 is a utility type for entry points that receive 0 parameters and return 1 value
type EP02[S isc.SandboxBase, R1, R2 any, O1 DictCodec[R1], O2 DictCodec[R2]] struct {
	EP0[S]
	Output1 O1
	Output2 O2
}

func NewViewEP02[R1, R2 any, O1 DictCodec[R1], O2 DictCodec[R2]](
	contract *ContractInfo, name string,
	out1 O1,
	out2 O2,
) EP02[isc.SandboxView, R1, R2, O1, O2] {
	return EP02[isc.SandboxView, R1, R2, O1, O2]{
		EP0:     NewViewEP0(contract, name),
		Output1: out1,
		Output2: out2,
	}
}

func (e EP02[S, R1, R2, O1, O2]) WithHandler(f func(ctx S) (R1, R2)) *EntryPointHandler[S] {
	return e.EntryPointInfo.WithHandler(func(ctx S) dict.Dict {
		r1, r2 := f(ctx)
		d := e.Output1.Encode(r1)
		d.Extend(e.Output2.Encode(r2))
		return d
	})
}

// EP11 is a utility type for entry points that receive 1 parameter and return 1 value
type EP11[S isc.SandboxView, T any, R any, I DictCodec[T], O DictCodec[R]] struct {
	EP1[S, T, I]
	Output O
}

func NewViewEP11[T any, R any, I DictCodec[T], O DictCodec[R]](
	contract *ContractInfo, name string,
	in I,
	out O,
) EP11[isc.SandboxView, T, R, I, O] {
	return EP11[isc.SandboxView, T, R, I, O]{
		EP1:    NewViewEP1(contract, name, in),
		Output: out,
	}
}

func (e EP11[S, T, R, I, O]) WithHandler(f func(S, T) R) *EntryPointHandler[S] {
	return e.EntryPointInfo.WithHandler(func(ctx S) dict.Dict {
		p, err := e.Input.Decode(ctx.Params().Dict)
		ctx.RequireNoError(err)
		r := f(ctx, p)
		return e.Output.Encode(r)
	})
}

// EP12 is a utility type for entry points that receive 1 parameter and return 1 value
type EP12[S isc.SandboxBase, T any, R1 any, R2 any, I DictCodec[T], O1 DictCodec[R1], O2 DictCodec[R2]] struct {
	EP1[S, T, I]
	Output1 O1
	Output2 O2
}

func NewViewEP12[T any, R1 any, R2 any, I DictCodec[T], O1 DictCodec[R1], O2 DictCodec[R2]](
	contract *ContractInfo, name string,
	in I,
	out1 O1,
	out2 O2,
) EP12[isc.SandboxView, T, R1, R2, I, O1, O2] {
	return EP12[isc.SandboxView, T, R1, R2, I, O1, O2]{
		EP1:     NewViewEP1(contract, name, in),
		Output1: out1,
		Output2: out2,
	}
}

func (e EP12[S, T, R1, R2, I, O1, O2]) WithHandler(f func(S, T) (R1, R2)) *EntryPointHandler[S] {
	return e.EntryPointInfo.WithHandler(func(ctx S) dict.Dict {
		p, err := e.Input.Decode(ctx.Params().Dict)
		ctx.RequireNoError(err)
		r1, r2 := f(ctx, p)
		d := e.Output1.Encode(r1)
		d.Extend(e.Output2.Encode(r2))
		return d
	})
}

// EP21 is a utility type for entry points that receive 2 parameters and return 1 value
type EP21[S isc.SandboxBase, T1 any, T2 any, R any, I1 DictCodec[T1], I2 DictCodec[T2], O DictCodec[R]] struct {
	EP2[S, T1, T2, I1, I2]
	Output O
}

func NewViewEP21[T1 any, T2 any, R any, I1 DictCodec[T1], I2 DictCodec[T2], O DictCodec[R]](
	contract *ContractInfo, name string,
	in1 I1,
	in2 I2,
	out O,
) EP21[isc.SandboxView, T1, T2, R, I1, I2, O] {
	return EP21[isc.SandboxView, T1, T2, R, I1, I2, O]{
		EP2:    NewViewEP2(contract, name, in1, in2),
		Output: out,
	}
}

func (e EP21[S, T1, T2, R, I1, I2, O]) WithHandler(f func(S, T1, T2) R) *EntryPointHandler[S] {
	return e.EntryPointInfo.WithHandler(func(ctx S) dict.Dict {
		params := ctx.Params()
		p1, err := e.Input1.Decode(params.Dict)
		ctx.RequireNoError(err)
		p2, err := e.Input2.Decode(params.Dict)
		ctx.RequireNoError(err)
		r := f(ctx, p1, p2)
		return e.Output.Encode(r)
	})
}