package vmimpl

import (
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
)

const MaxPostedOutputsInOneRequest = 4

func (vmctx *vmContext) getObject(chainState kv.KVStore, id iotago.ObjectID) (isc.L1Object, bool) {
	return vmctx.accountsStateWriterFromChainState(chainState).GetObject(id, vmctx.ChainID())
}

func (reqctx *requestContext) send(params isc.RequestParameters) {
	// simply send assets to a L1 address
	reqctx.vm.txbuilder.SendAssets(params.TargetAddress.AsIotaAddress(), params.Assets)
}
