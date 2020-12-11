package vmcontext

func (vmctx *VMContext) Infof(format string, params ...interface{}) {
	vmctx.log.Infof(format, params...)
}

func (vmctx *VMContext) Debugf(format string, params ...interface{}) {
	vmctx.log.Debugf(format, params...)
}

func (vmctx *VMContext) Panicf(format string, params ...interface{}) {
	vmctx.log.Panicf(format, params...)
}
