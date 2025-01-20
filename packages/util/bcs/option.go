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

	e.Encode(&p.Some)

	return nil
}

func (p *Option[T]) UnmarshalBCS(d *Decoder) error {
	p.None = !d.ReadOptionalFlag()
	if p.None {
		return nil
	}

	d.Decode(&p.Some)

	return nil
}
