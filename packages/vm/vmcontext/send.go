package vmcontext

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
)

func (vmctx *VMContext) getNFTData(nftID iotago.NFTID) *iscp.NFT {
	var nft *iscp.NFT
	vmctx.callCore(accounts.Contract, func(s kv.KVStore) {
		nft = accounts.GetNFTData(s, nftID)
	})
	return nft
}

func (vmctx *VMContext) SendAsNFT(par iscp.RequestParameters, nftID iotago.NFTID) {
	nft := vmctx.getNFTData(nftID)
	out := transaction.NFTOutputFromPostData(
		vmctx.task.AnchorOutput.AliasID.ToAddress(),
		vmctx.CurrentContractHname(),
		par,
		vmctx.task.L1Params.RentStructure(),
		nft,
	)
	vmctx.debitNFTFromAccount(vmctx.AccountID(), nftID)
	vmctx.sendOutput(out)
}

// Send implements sandbox function of sending cross-chain request
func (vmctx *VMContext) Send(par iscp.RequestParameters) {
	// create extended output
	out := transaction.BasicOutputFromPostData(
		vmctx.task.AnchorOutput.AliasID.ToAddress(),
		vmctx.CurrentContractHname(),
		par,
		vmctx.task.L1Params.RentStructure(),
	)
	vmctx.sendOutput(out)
}

func (vmctx *VMContext) sendOutput(o iotago.Output) {
	if vmctx.NumPostedOutputs >= MaxPostedOutputsInOneRequest {
		panic(vm.ErrExceededPostedOutputLimit)
	}
	vmctx.NumPostedOutputs++

	assets := iscp.FungibleTokensFromOutput(o)

	vmctx.assertConsistentL2WithL1TxBuilder("sandbox.Send: begin")
	// this call cannot panic due to not enough iotas for dust because
	// it does not change total balance of the transaction, and it does not create new internal outputs
	// The call can destroy internal output when all native tokens of particular ID are moved outside chain
	// The caller will receive all the dust deposit
	iotaAdjustmentL2 := vmctx.txbuilder.AddOutput(o)
	vmctx.adjustL2IotasIfNeeded(iotaAdjustmentL2, vmctx.AccountID())
	// debit the assets from the on-chain account
	// It panics with accounts.ErrNotEnoughFunds if sender's account balances are exceeded
	vmctx.debitFromAccount(vmctx.AccountID(), assets)
	vmctx.assertConsistentL2WithL1TxBuilder("sandbox.Send: end")
}
