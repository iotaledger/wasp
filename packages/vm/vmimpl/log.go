package vmimpl

import (
	"github.com/iotaledger/wasp/v2/packages/isc"
)

var _ isc.LogInterface = &requestContext{}

func (reqctx *requestContext) Infof(format string, params ...interface{}) {
	reqctx.vm.task.Log.LogInfof(format, params...)
}

func (reqctx *requestContext) Debugf(format string, params ...interface{}) {
	reqctx.vm.task.Log.LogDebugf(format, params...)
}

func (reqctx *requestContext) Panicf(format string, params ...interface{}) {
	reqctx.vm.task.Log.LogPanicf(format, params...)
}

func (reqctx *requestContext) Warnf(format string, params ...interface{}) {
	reqctx.vm.task.Log.LogWarnf(format, params...)
}
