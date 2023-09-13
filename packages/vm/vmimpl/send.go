package vmimpl

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
)

const MaxPostedOutputsInOneRequest = 4

func (vmctx *vmContext) getNFTData(chainState kv.KVStore, nftID iotago.NFTID) *isc.NFT {
	var nft *isc.NFT
	withContractState(chainState, accounts.Contract, func(s kv.KVStore) {
		nft = accounts.GetNFTData(s, nftID)
	})
	return nft
}

func (reqctx *requestContext) send(par isc.RequestParameters) {
	reqctx.doSend(isc.ContractIdentityFromHname(reqctx.CurrentContractHname()), par)
}

// Send implements sandbox function of sending cross-chain request
func (reqctx *requestContext) doSend(caller isc.ContractIdentity, par isc.RequestParameters) {
	if len(par.Assets.NFTs) > 1 {
		panic(vm.ErrSendMultipleNFTs)
	}
	if len(par.Assets.NFTs) == 1 {
		// create NFT output
		nft := reqctx.vm.getNFTData(reqctx.chainStateWithGasBurn(), par.Assets.NFTs[0])
		out := transaction.NFTOutputFromPostData(
			reqctx.vm.task.AnchorOutput.AliasID.ToAddress(),
			caller,
			par,
			nft,
		)
		debitNFTFromAccount(reqctx.chainStateWithGasBurn(), reqctx.CurrentContractAccountID(), nft.ID, reqctx.ChainID())
		reqctx.sendOutput(out)
		return
	}
	// create extended output
	out := transaction.BasicOutputFromPostData(
		reqctx.vm.task.AnchorOutput.AliasID.ToAddress(),
		caller,
		par,
	)
	reqctx.sendOutput(out)
}

func (reqctx *requestContext) sendOutput(o iotago.Output) {
	if reqctx.numPostedOutputs >= MaxPostedOutputsInOneRequest {
		panic(vm.ErrExceededPostedOutputLimit)
	}
	reqctx.numPostedOutputs++

	assets := isc.AssetsFromOutput(o)

	// this call cannot panic due to not enough base tokens for storage deposit because
	// it does not change total balance of the transaction, and it does not create new internal outputs
	// The call can destroy internal output when all native tokens of particular ID are moved outside chain
	// The caller will receive all the storage deposit
	baseTokenAdjustmentL2 := reqctx.vm.txbuilder.AddOutput(o)
	reqctx.adjustL2BaseTokensIfNeeded(baseTokenAdjustmentL2, reqctx.CurrentContractAccountID())
	// debit the assets from the on-chain account
	// It panics with accounts.ErrNotEnoughFunds if sender's account balances are exceeded
	debitFromAccount(reqctx.chainStateWithGasBurn(), reqctx.CurrentContractAccountID(), assets, reqctx.ChainID())
}
