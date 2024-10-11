package codec

import (
	"bytes"
	"fmt"
	"io"

	"github.com/samber/lo"

	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/util/bcs"
)

type Codec[T any] interface {
	Encode(T) []byte
	Decode([]byte, ...T) (T, error)
	MustDecode([]byte, ...T) T

	EncodeStream(w io.Writer, v T)
	DecodeStream(r io.Reader) (T, error)
	MustDecodeStream(r io.Reader) T
}

type codec[T any] struct {
	decode func(r io.Reader) (T, error)
	encode func(w io.Writer, v T)
}

func NewCodec[T any](decode func(r io.Reader) (T, error), encode func(w io.Writer, v T)) Codec[T] {
	return &codec[T]{decode: decode, encode: encode}
}

func NewCodecFromBCS[T any]() Codec[T] {
	encode := func(w io.Writer, v T) { bcs.MustMarshalStream(&v, w) }
	decode := bcs.UnmarshalStream[T]
	return &codec[T]{decode: decode, encode: encode}
}

func NewTupleCodec[
	A, B any,
]() Codec[lo.Tuple2[A, B]] {
	return NewCodecFromBCS[lo.Tuple2[A, B]]()
}

func (c *codec[T]) Decode(b []byte, def ...T) (v T, err error) {
	if b == nil {
		if len(def) == 0 {
			err = fmt.Errorf("%T: cannot decode nil bytes", v)
			return
		}
		return def[0], nil
	}

	r := bytes.NewReader(b)
	v, err = c.decode(r)
	if err != nil {
		return v, fmt.Errorf("%T: %w", v, err)
	}

	if r.Len() > 0 {
		return v, fmt.Errorf("%T: %v bytes left after decoding", v, r.Len())
	}

	return v, nil
}

func (c *codec[T]) MustDecode(b []byte, def ...T) (r T) {
	return lo.Must(c.Decode(b, def...))
}

func (c *codec[T]) Encode(v T) []byte {
	var buf bytes.Buffer
	c.encode(&buf, v)

	return buf.Bytes()
}

func (c *codec[T]) EncodeStream(w io.Writer, v T) {
	c.encode(w, v)
}

func (c *codec[T]) DecodeStream(r io.Reader) (T, error) {
	v, err := c.decode(r)
	if err != nil {
		return v, fmt.Errorf("%T: %w", v, err)
	}

	return v, nil
}

func (c *codec[T]) MustDecodeStream(r io.Reader) T {
	return lo.Must(c.DecodeStream(r))
}

func SliceToArray[T any](c Codec[T], slice []T) []byte {
	e := bcs.NewBytesEncoder()

	e.WriteLen(len(slice))

	for _, v := range slice {
		c.EncodeStream(e, v)
	}

	if err := e.Err(); err != nil {
		panic(err)
	}

	return e.Bytes()
}

func SliceFromArray[T any](c Codec[T], b []byte) ([]T, error) {
	if len(b) == 0 {
		return nil, nil
	}

	d := bcs.NewBytesDecoder(b)
	length := d.ReadLen()

	ret := make([]T, length)

	for i := range ret {
		var err error
		ret[i], err = c.DecodeStream(d)
		if err != nil {
			return nil, err
		}
	}

	if err := d.Err(); err != nil {
		return nil, err
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
