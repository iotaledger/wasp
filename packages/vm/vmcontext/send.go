package vmcontext

import (
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/vm/vmcontext/vmtxbuilder"
)

// Send implements sandbox function of sending cross-chain request
func (vmctx *VMContext) Send(par iscp.RequestParameters) {
	// create extended output with adjusted dust deposit
	out := vmtxbuilder.ExtendedOutputFromPostData(
		vmctx.task.AnchorOutput.AliasID.ToAddress(),
		vmctx.CurrentContractHname(),
		par,
		vmctx.task.RentStructure,
	)
	// the output must have exactly the amount of assets specified y the used
	// If it is different, it means some iotas were added to fulfill dust requirements. Not good
	if !par.Assets.Equals(iscp.AssetsFromOutput(out)) {
		panic(ErrNotEnoughIotasForDustDeposit)
	}
	// debit the assets from the on-chain account
	// It panics with accounts.ErrNotEnoughFunds if sender's account balances are exceeded
	vmctx.debitFromAccount(vmctx.AccountID(), par.Assets)
	// this call cannot panic due to not enough iotas for dust because
	// it does not change total balance of the transaction and it does not create new internal outputs
	// The call can destroy internal output when all native tokens of particular ID are moved outside chain
	vmctx.txbuilder.AddOutput(out)
	vmctx.assertConsistentL2WithL1TxBuilder("sandbox.Send: end")
}
