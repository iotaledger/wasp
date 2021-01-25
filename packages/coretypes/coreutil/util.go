package coreutil

import (
	"fmt"
	"github.com/iotaledger/wasp/packages/coretypes"
)

type assert struct {
	log coretypes.LogInterface
}

func NewAssert(log ...coretypes.LogInterface) assert {
	if len(log) == 0 {
		return assert{}
	}
	return assert{log: log[0]}
}

func (a assert) Require(cond bool, format string, args ...interface{}) {
	if cond {
		return
	}
	if a.log == nil {
		panic(fmt.Sprintf(format, args...))
	}
	a.log.Panicf(format, args...)
}

func (a assert) RequireNoError(err error) {
	a.Require(err == nil, fmt.Sprintf("%v", err))
}
