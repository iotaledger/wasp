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
	// TODO: Here used to be Cross-Chain logic.
	// Look at ac543caf8ddb276d5580c3ef5d4a41c6ef90e625 to see how it was implemented in the past.
	// For now removed everything related to it.

	// simply send assets to a L1 address
	reqctx.vm.txbuilder.SendAssets(params.TargetAddress.AsIotaAddress(), params.Assets)
}
