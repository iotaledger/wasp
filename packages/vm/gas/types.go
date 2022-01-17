package gas

import "golang.org/x/xerrors"

type BurnCode uint16

type BurnFunction func(code BurnCode, params []int) uint64

type BurnCodeRecord struct {
	Name string
	BurnFunction
}

type BurnTable map[BurnCode]BurnCodeRecord

var (
	ErrUnknownBurnCode            = xerrors.New("unknown gas burn code")
	ErrInLinear1ParameterExpected = xerrors.New("'linear' gas burn requires exactly 1 parameter")
)

func (c BurnCode) Name() string {
	r, ok := burnTable[c]
	if !ok {
		return "(undef)"
	}
	return r.Name
}

func Value(c BurnCode, p ...int) uint64 {
	if r, ok := burnTable[c]; ok {
		return r.BurnFunction(c, p)
	}
	panic(xerrors.Errorf("%v: %d", ErrUnknownBurnCode, c))

}
