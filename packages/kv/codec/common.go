package codec

import (
	"fmt"

	"github.com/samber/lo"
)

type Codec[T any] struct {
	decode func([]byte) (T, error)
	encode func(T) []byte
}

func NewCodec[T any](decode func([]byte) (T, error), encode func(T) []byte) *Codec[T] {
	return &Codec[T]{decode: decode, encode: encode}
}

func NewCodecEx[T interface{ Bytes() []byte }](decode func([]byte) (T, error)) *Codec[T] {
	return &Codec[T]{decode: decode, encode: func(v T) []byte {
		return v.Bytes()
	}}
}

func (c *Codec[T]) Decode(b []byte, def ...T) (r T, err error) {
	if b == nil {
		if len(def) == 0 {
			err = fmt.Errorf("%T: cannot decode nil bytes", r)
			return
		}
		return def[0], nil
	}
	return c.decode(b)
}

func (c *Codec[T]) MustDecode(b []byte, def ...T) (r T) {
	return lo.Must(c.Decode(b, def...))
}

func (c *Codec[T]) Encode(v T) []byte {
	return c.encode(v)
}
