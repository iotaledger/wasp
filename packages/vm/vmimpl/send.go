package vmimpl

import (
	"github.com/iotaledger/wasp/clients/iota-go/iotago"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
)

const MaxPostedOutputsInOneRequest = 4

func (vmctx *vmContext) getNFTData(chainState kv.KVStore, nftID iotago.ObjectID) *isc.NFT {
	panic("refactor me: getNFTData vm.send")
	// return vmctx.accountsStateWriterFromChainState(chainState).GetNFTData(nftID)
}

func (reqctx *requestContext) send(params isc.RequestParameters) {
	// simply send assets to a L1 address
	reqctx.vm.txbuilder.SendAssets(params.TargetAddress.AsIotaAddress(), params.Assets)
}
