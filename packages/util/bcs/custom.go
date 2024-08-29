package bcs

import (
	"math/big"
	"time"
)

func init() {
	AddCustomEncoder(func(e *Encoder, v big.Int) error {
		return EncodeUint128(&v, e)
	})

	AddCustomDecoder(func(d *Decoder) (big.Int, error) {
		v, err := DecodeUint128(d)
		if err != nil {
			return big.Int{}, err
		}

		return *v, nil
	})

	AddCustomEncoder(func(e *Encoder, v time.Time) error {
		e.w.WriteInt64(v.UnixNano())
		return e.w.Err
	})

	AddCustomDecoder(func(d *Decoder) (time.Time, error) {
		v := time.Unix(0, d.r.ReadInt64())
		return v, d.r.Err
	})
}
