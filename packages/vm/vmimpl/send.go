package vmimpl

import (
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/sui-go/sui"
)

const MaxPostedOutputsInOneRequest = 4

func (vmctx *vmContext) getNFTData(chainState kv.KVStore, nftID sui.ObjectID) *isc.NFT {
	panic("refactor me: getNFTData vm.send")
	//return vmctx.accountsStateWriterFromChainState(chainState).GetNFTData(nftID)
}

func (reqctx *requestContext) send(params isc.RequestParameters) {
	if params.Metadata == nil {
		// simply send assets to a L1 address
		reqctx.vm.txbuilder.SendAssets(params.TargetAddress.AsSuiAddress(), params.Assets)
	} else {
		// sending cross chain request to a contract on the other chain
		reqctx.vm.txbuilder.SendCrossChainRequest(params.TargetAddress.AsSuiAddress(), reqctx.vm.task.Anchor.Ref.ObjectID, params.Assets, params.Metadata)
	}
}
