package vmcontext

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/vm/vmcontext/vmtxbuilder"
	"golang.org/x/xerrors"
)

// Send implements sandbox function of sending cross-chain request
func (vmctx *VMContext) Send(target iotago.Address, assets *iscp.Assets, sendMetadata *iscp.SendMetadata, options ...*iscp.SendOptions) {
	if assets == nil {
		panic(xerrors.New("post request assets can't be nil"))
	}
	// create extended output with adjusted dust deposit
	extendedOutput := vmtxbuilder.ExtendedOutputFromPostData(
		target,
		vmctx.task.AnchorOutput.AliasID.ToAddress(),
		vmctx.CurrentContractHname(),
		assets,
		sendMetadata,
		options...,
	)
	// debit the assets from the on-chain account
	vmctx.debitFromAccount(vmctx.AccountID(), &iscp.Assets{
		Iotas:  extendedOutput.Amount,
		Tokens: extendedOutput.NativeTokens,
	})
	vmctx.txbuilder.AddOutput(extendedOutput)
	// TODO check consistency between transaction builder and the on-chain accounts
}
