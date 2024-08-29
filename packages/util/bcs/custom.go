package bcs

import (
	"math/big"
	"time"
)

func init() {
	// AddCustomEncoder(func(e *Encoder, v *big.Int) error {
	// 	e.w.WriteBigUint(v)
	// 	return e.w.Err
	// })

	AddCustomEncoder(func(e *Encoder, v big.Int) error {
		e.w.WriteBigUint(&v)
		return e.w.Err
	})

	// AddCustomDecoder(func(d *Decoder) (*big.Int, error) {
	// 	v := d.r.ReadBigUint()
	// 	return v, d.r.Err
	// })

	AddCustomDecoder(func(d *Decoder) (big.Int, error) {
		v := d.r.ReadBigUint()
		return *v, d.r.Err
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
