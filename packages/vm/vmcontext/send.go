package vmcontext

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
)

func (vmctx *VMContext) getNFTData(nftID iotago.NFTID) *isc.NFT {
	var nft isc.NFT
	vmctx.callCore(accounts.Contract, func(s kv.KVStore) {
		nft = accounts.GetNFTData(s, nftID)
	})
	return &nft
}

func (vmctx *VMContext) SendAsNFT(par isc.RequestParameters, nftID iotago.NFTID) {
	nft := vmctx.getNFTData(nftID)
	out := transaction.NFTOutputFromPostData(
		vmctx.task.AnchorOutput.AliasID.ToAddress(),
		vmctx.CurrentContractHname(),
		par,
		nft,
	)
	vmctx.debitNFTFromAccount(vmctx.AccountID(), nftID)
	vmctx.sendOutput(out)
}

// Send implements sandbox function of sending cross-chain request
func (vmctx *VMContext) Send(par isc.RequestParameters) {
	// create extended output
	out := transaction.BasicOutputFromPostData(
		vmctx.task.AnchorOutput.AliasID.ToAddress(),
		vmctx.CurrentContractHname(),
		par,
	)
	vmctx.sendOutput(out)
}

func (vmctx *VMContext) sendOutput(o iotago.Output) {
	if vmctx.NumPostedOutputs >= MaxPostedOutputsInOneRequest {
		panic(vm.ErrExceededPostedOutputLimit)
	}
	vmctx.NumPostedOutputs++

	assets := isc.FungibleTokensFromOutput(o)

	vmctx.assertConsistentL2WithL1TxBuilder("sandbox.Send: begin")
	// this call cannot panic due to not enough base tokens for storage deposit because
	// it does not change total balance of the transaction, and it does not create new internal outputs
	// The call can destroy internal output when all native tokens of particular ID are moved outside chain
	// The caller will receive all the storage deposit
	baseTokenAdjustmentL2 := vmctx.txbuilder.AddOutput(o)
	vmctx.adjustL2BaseTokensIfNeeded(baseTokenAdjustmentL2, vmctx.AccountID())
	// debit the assets from the on-chain account
	// It panics with accounts.ErrNotEnoughFunds if sender's account balances are exceeded
	vmctx.debitFromAccount(vmctx.AccountID(), assets)
	vmctx.assertConsistentL2WithL1TxBuilder("sandbox.Send: end")
}
