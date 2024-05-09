package codec

import (
	"fmt"

	"github.com/samber/lo"

	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/collections"
	"github.com/iotaledger/wasp/packages/kv/dict"
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

func SliceToArray[T any](c Codec[T], slice []T, arrayKey string) dict.Dict {
	ret := dict.Dict{}
	retArr := collections.NewArray(ret, arrayKey)
	for _, v := range slice {
		retArr.Push(c.Encode(v))
	}
	return ret
}

func SliceFromArray[T any](c Codec[T], d dict.Dict, arrayKey string) ([]T, error) {
	if len(d) == 0 {
		return nil, nil
	}
	arr := collections.NewArrayReadOnly(d, arrayKey)
	ret := make([]T, arr.Len())
	for i := range ret {
		var err error
		ret[i], err = c.Decode(arr.GetAt(uint32(i)))
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
