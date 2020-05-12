package variables

import "io"

type Variables interface {
	// TODO tbd
	// interface to handle variable/value pairs
	// used in variable state, state updates and requests
	Read(io.Reader) error
	Write(io.Writer) error
}

func NewVariables() Variables {
	panic("implement me")
}
