package vmcontext

// getBlockContext returns object of the block context or nil if it was not created
func (vmctx *VMContext) GetBlockContext() (interface{}, bool) {
	ret, ok := vmctx.blockContext[vmctx.CurrentContractHname()]
	return ret, ok
}

func (vmctx *VMContext) CreateBlockContext(obj interface{}, onClose ...func()) {
	hname := vmctx.CurrentContractHname()
	if _, alreadyExists := vmctx.blockContext[hname]; alreadyExists {
		vmctx.log.Panicf("attempt to create block context twice")
	}
	fun := func() {}
	if len(onClose) > 0 {
		fun = onClose[0]
	}
	vmctx.blockContext[hname] = &blockContext{
		obj:     obj,
		onClose: fun,
	}
	// storing sequence to have deterministic order of closing
	vmctx.blockContextCloseSeq = append(vmctx.blockContextCloseSeq, hname)
}
