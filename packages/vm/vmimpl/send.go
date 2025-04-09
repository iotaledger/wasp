package vmimpl

import (
	"github.com/iotaledger/wasp/packages/isc"
)

const MaxPostedOutputsInOneRequest = 4

func (reqctx *requestContext) send(params isc.RequestParameters) {
	// simply send assets to a L1 address
	reqctx.vm.txbuilder.SendAssets(params.TargetAddress.AsIotaAddress(), params.Assets)
}
