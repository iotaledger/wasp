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
	if params.Metadata == nil {
		// simply send assets to a L1 address
		reqctx.vm.txbuilder.SendAssets(params.TargetAddress.AsIotaAddress(), params.Assets)
	} else {
		// sending cross chain request to a contract on the other chain
		reqctx.vm.txbuilder.SendCrossChainRequest(
			&reqctx.vm.task.Anchor.ISCPackage,
			params.TargetAddress.AsIotaAddress(),
			params.Assets,
			params.Metadata,
		)
	}
}
