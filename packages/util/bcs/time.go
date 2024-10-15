package bcs

import (
	"time"
)

func init() {
	AddCustomEncoder(func(e *Encoder, v time.Time) error {
		var ns int64
		if !v.IsZero() {
			ns = v.UnixNano()
		}

		e.w.WriteInt64(ns)
		return e.w.Err
	})

	AddCustomDecoder(func(d *Decoder, v *time.Time) error {
		ns := d.r.ReadInt64()
		if d.r.Err != nil {
			return d.r.Err
		}

		if ns == 0 {
			*v = time.Time{}
			return nil
		}

		*v = time.Unix(0, ns)

		return nil
	})
}
