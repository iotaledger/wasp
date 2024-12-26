package bcs

type Option[T any] struct {
	Some T
	None bool
}

func (p *Option[T]) MarshalBCS(e *Encoder) error {
	err := e.WriteOptionalFlag(!p.None)
	if err != nil {
		return err
	}
	if p.None {
		return nil
	}
	return e.Encode(&p.Some)
}

func (p *Option[T]) UnmarshalBCS(d *Decoder) error {
	p.None = !d.ReadOptionalFlag()
	if d.Err() != nil {
		return d.Err()
	}
	if p.None {
		return nil
	}
	return d.Decode(&p.Some)
}
