package bcs

import "time"

func init() {
	AddCustomEncoder(func(e *Encoder, v time.Time) error {
		e.w.WriteInt64(v.UnixNano())
		return e.w.Err
	})

	AddCustomDecoder(func(d *Decoder, v *time.Time) error {
		*v = time.Unix(0, d.r.ReadInt64())
		return d.r.Err
	})
}
