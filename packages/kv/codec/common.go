package codec

import (
	"fmt"

	"github.com/samber/lo"

	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/util/rwutil"
)

type Codec[T any] interface {
	Encode(T) []byte
	Decode([]byte, ...T) (T, error)
	MustDecode([]byte, ...T) T
}

type codec[T any] struct {
	decode func([]byte) (T, error)
	encode func(T) []byte
}

func NewCodec[T any](decode func([]byte) (T, error), encode func(T) []byte) Codec[T] {
	return &codec[T]{decode: decode, encode: encode}
}

func NewCodecEx[T interface{ Bytes() []byte }](decode func([]byte) (T, error)) Codec[T] {
	return &codec[T]{decode: decode, encode: func(v T) []byte {
		return v.Bytes()
	}}
}

func NewCodecFromIoReadWriter[T rwutil.IoReadWriter]() Codec[T] {
	encode := func(obj T) []byte { return rwutil.WriteToBytes(obj) }
	decode := func(b []byte) (T, error) {
		return rwutil.ReadFromBytes(b, rwutil.MakeValue[T]())
	}
	return &codec[T]{decode: decode, encode: encode}
}

func NewTupleCodec[
	A, B rwutil.IoReadWriter,
]() Codec[lo.Tuple2[A, B]] {
	encode := func(obj lo.Tuple2[A, B]) []byte {
		ww := rwutil.NewBytesWriter()
		ww.Write(obj.A)
		ww.Write(obj.B)
		return ww.Bytes()
	}
	decode := func(b []byte) (lo.Tuple2[A, B], error) {
		rr := rwutil.NewBytesReader(b)
		ret := lo.Tuple2[A, B]{
			A: rwutil.MakeValue[A](),
			B: rwutil.MakeValue[B](),
		}
		rr.Read(ret.A)
		rr.Read(ret.B)
		return ret, rr.Err
	}
	return NewCodec(decode, encode)
}

func (c *codec[T]) Decode(b []byte, def ...T) (r T, err error) {
	if b == nil {
		if len(def) == 0 {
			err = fmt.Errorf("%T: cannot decode nil bytes", r)
			return
		}
		return def[0], nil
	}
	return c.decode(b)
}

func (c *codec[T]) MustDecode(b []byte, def ...T) (r T) {
	return lo.Must(c.Decode(b, def...))
}

func (c *codec[T]) Encode(v T) []byte {
	return c.encode(v)
}

func SliceToArray[T any](c Codec[T], slice []T) []byte {
	w := rwutil.NewBytesWriter()
	w.WriteSize32(len(slice))
	for _, v := range slice {
		value := c.Encode(v)
		w.WriteBytes(value)
	}
	return w.Bytes()
}

func SliceFromArray[T any](c Codec[T], d []byte) ([]T, error) {
	if len(d) == 0 {
		return nil, nil
	}

	r := rwutil.NewBytesReader(d)
	length := r.ReadSize32()

	ret := make([]T, length)

	for i := range ret {
		var err error
		ret[i], err = c.Decode(r.ReadBytes())
		if err != nil {
			return nil, err
		}
	}

	return ret, nil
}

func SliceToDictKeys[T any](c Codec[T], set []T) dict.Dict {
	ret := dict.Dict{}
	for _, v := range set {
		ret[kv.Key(c.Encode(v))] = []byte{0x01}
	}
	return ret
}

func SliceFromDictKeys[T any](c Codec[T], r dict.Dict) ([]T, error) {
	ret := make([]T, 0, len(r))
	for k := range r {
		v, err := c.Decode([]byte(k))
		if err != nil {
			return nil, err
		}
		ret = append(ret, v)
	}
	return ret, nil
}
