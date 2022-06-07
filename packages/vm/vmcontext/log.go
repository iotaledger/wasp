package vmcontext

import (
	"github.com/iotaledger/wasp/packages/iscp"
)

var _ iscp.LogInterface = &VMContext{}

func (vmctx *VMContext) Infof(format string, params ...interface{}) {
	vmctx.task.Log.Infof(format, params...)
}

func (vmctx *VMContext) Debugf(format string, params ...interface{}) {
	vmctx.task.Log.Debugf(format, params...)
}

func (vmctx *VMContext) Panicf(format string, params ...interface{}) {
	vmctx.task.Log.Panicf(format, params...)
}

func (vmctx *VMContext) Warnf(format string, params ...interface{}) {
	vmctx.task.Log.Warnf(format, params...)
}
