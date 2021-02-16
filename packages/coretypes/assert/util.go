package assert

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/coretypes"
)

type Assert struct {
	log coretypes.LogInterface
}

func NewAssert(log ...coretypes.LogInterface) Assert {
	if len(log) == 0 {
		return Assert{}
	}
	return Assert{log: log[0]}
}

func (a Assert) Require(cond bool, format string, args ...interface{}) {
	if cond {
		return
	}
	if a.log == nil {
		panic(fmt.Sprintf(format, args...))
	}
	a.log.Panicf(format, args...)
}

func (a Assert) RequireNoError(err error) {
	a.Require(err == nil, fmt.Sprintf("%v", err))
}
