package bcs

type Option[T any] struct {
	Some T
	None bool
}

func (p *Option[T]) MarshalBCS(e *Encoder) error {
	e.WriteOptionalFlag(!p.None)

	if p.None {
		return nil
	}

	return e.Encode(&p.Some)
}

func (p *Option[T]) UnmarshalBCS(d *Decoder) error {
	p.None = !d.ReadOptionalFlag()
	if p.None {
		return nil
	}

	return d.Decode(&p.Some)
}
