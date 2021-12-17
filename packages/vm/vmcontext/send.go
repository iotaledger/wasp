package vmcontext

import (
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/vm/vmcontext/vmtxbuilder"
	"golang.org/x/xerrors"
)

// Send implements sandbox function of sending cross-chain request
func (vmctx *VMContext) Send(par iscp.RequestParameters) bool {
	if par.Assets == nil {
		panic(xerrors.New("post request assets can't be nil"))
	}
	// create extended output with adjusted dust deposit
	extendedOutput := vmtxbuilder.ExtendedOutputFromPostData(
		vmctx.task.AnchorOutput.AliasID.ToAddress(),
		vmctx.CurrentContractHname(),
		par,
		vmctx.task.RentStructure,
	)
	// debit the assets from the on-chain account
	// It panics with accounts.ErrNotEnoughFunds if sender's account balances are exceeded
	vmctx.debitFromAccount(vmctx.AccountID(), &iscp.Assets{
		Iotas:  extendedOutput.Amount,
		Tokens: extendedOutput.NativeTokens,
	})
	// this call cannot panic due to not enough iotas for dust because
	// it does not change total balance of the transaction and it does not create new internal outputs
	// The call can destroy internal output when all native tokens of particular ID are moved outside chain
	vmctx.txbuilder.AddOutput(extendedOutput)
	// TODO check consistency between transaction builder and the on-chain accounts
	return true
}
