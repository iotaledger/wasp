package wbf

import (
	"math/big"
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
}
