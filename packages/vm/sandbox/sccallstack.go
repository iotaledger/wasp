package sandbox

import "github.com/iotaledger/wasp/packages/coretypes"

// mustCurrent can panic
func (vctx *sandbox) mustCurrentContractIndex() coretypes.Uint16 {
	return vctx.contractCallStack[len(vctx.contractCallStack)]
}

func (vctx *sandbox) pushContractIndex(cindex coretypes.Uint16) {
	vctx.contractCallStack = append(vctx.contractCallStack, cindex)
}

// mustPopContractIndex may panic
func (vctx *sandbox) mustPopContractIndex() {
	vctx.contractCallStack = vctx.contractCallStack[:len(vctx.contractCallStack)-1]
}
