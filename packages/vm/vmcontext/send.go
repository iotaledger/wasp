package vmcontext

import (
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/vm/gas"
	"golang.org/x/xerrors"
)

const MaxPostedOutputsInOneRequest = 4

// Send implements sandbox function of sending cross-chain request
func (vmctx *VMContext) Send(par iscp.RequestParameters) {
	if par.Assets == nil {
		panic(ErrAssetsCantBeEmptyInSend)
	}
	if vmctx.numPostedOutputs >= MaxPostedOutputsInOneRequest {
		panic(xerrors.Errorf("%v: max = %d", ErrExceededPostedOutputLimit, MaxPostedOutputsInOneRequest))
	}

	vmctx.numPostedOutputs++
	vmctx.GasBurn(gas.BurnSendL1Request, vmctx.numPostedOutputs)

	// create extended output with adjusted dust deposit
	out, err := transaction.ExtendedOutputFromPostData(
		vmctx.task.AnchorOutput.AliasID.ToAddress(),
		vmctx.CurrentContractHname(),
		par,
		vmctx.task.RentStructure,
	)
	if err != nil {
		// only possible if not provided enough iotas for dust deposit
		panic(err)
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
