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

		e.WriteInt64(ns)
		return e.err
	})

	AddCustomDecoder(func(d *Decoder, v *time.Time) error {
		ns := d.ReadInt64()
		if d.err != nil {
			return d.err
		}

		if ns == 0 {
			*v = time.Time{}
			return nil
		}

		*v = time.Unix(0, ns)

		return nil
	})
}
