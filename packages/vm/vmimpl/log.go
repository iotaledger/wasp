package vmimpl

import (
	"github.com/iotaledger/wasp/packages/isc"
)

var _ isc.LogInterface = &vmContext{}

func (vmctx *vmContext) Infof(format string, params ...interface{}) {
	vmctx.task.Log.Infof(format, params...)
}

func (vmctx *vmContext) Debugf(format string, params ...interface{}) {
	vmctx.task.Log.Debugf(format, params...)
}

func (vmctx *vmContext) Panicf(format string, params ...interface{}) {
	vmctx.task.Log.Panicf(format, params...)
}

func (vmctx *vmContext) Warnf(format string, params ...interface{}) {
	vmctx.task.Log.Warnf(format, params...)
}
