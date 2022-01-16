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
	ErrUnknownBurnCode = xerrors.New("unknown gas burn code")
)

func constValue(constGas uint64) BurnFunction {
	g := constGas
	return func(_ BurnCode, _ []int) uint64 {
		return g
	}
}

func notImplemented() BurnFunction {
	return func(code BurnCode, _ []int) uint64 {
		panic(xerrors.Errorf("burn code %d not implemented", code))
	}
}

func (c BurnCode) Value(p ...int) uint64 {
	return Value(c, p...)
}

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
