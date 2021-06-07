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

func (a Assert) RequireChainOwner(ctx coretypes.Sandbox, name ...string) {
	if !ctx.ChainOwnerID().Equals(ctx.Caller()) {
		if len(name) > 0 {
			a.log.Panicf("%s: unauthorized access", name[0])
		} else {
			a.log.Panicf("unauthorized access")
		}
	}
}
