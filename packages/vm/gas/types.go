package gas

import "errors"

type BurnCode uint16

type BurnFunction func(x uint64) uint64

type BurnCodeRecord struct {
	Name string
	BurnFunction
}

type BurnTable map[BurnCode]BurnCodeRecord

var ErrUnknownBurnCode = errors.New("unknown gas burn code")

func (c BurnCode) Name() string {
	r, ok := burnTable[c]
	if !ok {
		return "(undef)"
	}
	return r.Name
}
