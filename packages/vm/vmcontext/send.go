package vmcontext

import (
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/vmcontext/vmtxbuilder"
	"golang.org/x/xerrors"
)

// Send implements sandbox function of sending cross-chain request
func (vmctx *VMContext) Send(par iscp.RequestParameters) {
	if par.Assets == nil {
		panic(ErrAssetsCantBeEmptyInSend)
	}
	// create extended output with adjusted dust deposit
	out := vmtxbuilder.ExtendedOutputFromPostData(
		vmctx.task.AnchorOutput.AliasID.ToAddress(),
		vmctx.CurrentContractHname(),
		par,
		vmctx.task.RentStructure,
	)
	// the output must have exactly the amount of iotas specified by the user
	// If it is different it means some iotas were added to adjust for dust requirements. Not good
	if par.Assets.Iotas != out.Amount {
		panic(xerrors.Errorf("Send: %v: needed at least %d, got %d",
			accounts.ErrNotEnoughIotasForDustDeposit, out.Amount, par.Assets.Iotas))
	}
	vmctx.assertConsistentL2WithL1TxBuilder("sandbox.Send: begin")
	// this call cannot panic due to not enough iotas for dust because
	// it does not change total balance of the transaction and it does not create new internal outputs
	// The call can destroy internal output when all native tokens of particular ID are moved outside chain
	iotaAdjustmentL2 := vmctx.txbuilder.AddOutput(out)
	vmctx.adjustL2IotasIfNeeded(iotaAdjustmentL2)
	// debit the assets from the on-chain account
	// It panics with accounts.ErrNotEnoughFunds if sender's account balances are exceeded
	vmctx.debitFromAccount(vmctx.AccountID(), par.Assets)
	vmctx.assertConsistentL2WithL1TxBuilder("sandbox.Send: end")
}
