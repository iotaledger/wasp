package codec

import (
	"bytes"
	"fmt"
	"io"

	"github.com/samber/lo"

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

func Decode[T any](b []byte, def ...T) (v T, err error) {
	if b == nil {
		if len(def) == 0 {
			err = fmt.Errorf("%T: cannot decode nil bytes", v)
			return
		}
		return def[0], nil
	}

	r := bytes.NewReader(b)
	v, err = bcs.UnmarshalStream[T](r)
	if err != nil {
		return v, fmt.Errorf("%T: %w", v, err)
	}

	if r.Len() > 0 {
		return v, fmt.Errorf("%T: %v bytes left after decoding", v, r.Len())
	}

	return v, nil
}

func MustDecode[T any](b []byte, def ...T) (r T) {
	return lo.Must(Decode(b, def...))
}

func DecodeOptional[T any](b []byte) (v *T, err error) {
	o, err := bcs.Unmarshal[bcs.Option[*T]](b)
	if err != nil {
		return nil, fmt.Errorf("%T: %w", v, err)
	}

	return o.Some, nil
}

func Encode[T any](v T) []byte {
	return bcs.MustMarshal(&v)
}

func EncodeOptional[T any](v *T) []byte {
	o := bcs.Option[*T]{}

	if v != nil {
		o.Some = v
	} else {
		o.None = true
	}

	return bcs.MustMarshal(&o)
}
