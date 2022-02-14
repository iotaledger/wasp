package runvm

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/core/errors/commonerrors"
	"github.com/iotaledger/wasp/packages/vm/vmcontext"
	"github.com/iotaledger/wasp/packages/vm/vmerrors"
)

var ErrUndefinedError = commonerrors.RegisterGlobalError("&%v")

type VMRunner struct{}

func (r VMRunner) Run(task *vm.VMTask) {
	// optimistic read panic catcher for the whole VM task
	err := util.CatchPanicReturnError(
		func() { runTask(task) },
		coreutil.ErrorStateInvalidated,
	)
	if err != nil {

		switch e := err.(type) {
		case *vmerrors.Error:
			task.VMError = e
		case error:
			// May require a different error type here?
			task.VMError = ErrUndefinedError.Create(e)
		default:
			task.VMError = ErrUndefinedError.Create(e)
		}

		task.Log.Warnf("VM task has been abandoned due to invalidated state. ACS session id: %d", task.ACSSessionID)
	}
}

func NewVMRunner() vm.VMRunner {
	return VMRunner{}
}

// runTask runs batch of requests on VM
func runTask(task *vm.VMTask) {
	vmctx := vmcontext.CreateVMContext(task)

	var numOffLedger, numSuccess uint16
	reqIndexInTheBlock := 0

	// main loop over the batch of requests
	for _, req := range task.Requests {
		result, skipReason := vmctx.RunTheRequest(req, uint16(reqIndexInTheBlock))
		if skipReason != nil {
			// some requests are just ignored (deterministically)
			task.Log.Warnf("request skipped (ignored) by the VM: %s, reason: %v",
				req.ID().String(), skipReason)
			continue
		}
		task.Results = append(task.Results, result)
		reqIndexInTheBlock++
		if req.IsOffLedger() {
			numOffLedger++
		}

		if result.Error == nil {
			numSuccess++
		} else {
			task.Log.Debugf("runTask, ERROR running request: %s, error: %v", req.ID().String(), result.Error)
		}
		vmctx.AssertConsistentGasTotals()
	}

	numProcessed := uint16(len(task.Results))

	task.Log.Debugf("runTask, ran %d requests. success: %d, offledger: %d",
		numProcessed, numSuccess, numOffLedger)

	blockIndex, stateCommitment, timestamp, rotationAddr := vmctx.CloseVMContext(
		numProcessed, numSuccess, numOffLedger)

	task.Log.Debugf("closed VMContext: block index: %d, state hash: %s timestamp: %v, rotationAddr: %v",
		blockIndex, stateCommitment, timestamp, rotationAddr)

	if rotationAddr == nil {
		// rotation does not happen
		task.ResultTransactionEssence = vmctx.BuildTransactionEssence(&iscp.StateData{
			Commitment: stateCommitment,
		})

		// TODO extract latest total assets
		checkTotalAssets(task.ResultTransactionEssence, nil)

		task.Log.Debugf("runTask OUT. block index: %d, state hash: %s",
			blockIndex, stateCommitment.String(),
			//" tx essence hash: ", hashing.HashData(task.ResultTransactionEssence.Bytes()).String(),
		)
	} else {
		// rotation happens
		task.RotationAddress = rotationAddr
		task.ResultTransactionEssence = nil
		task.Log.Debugf("runTask OUT: rotate to address %s", rotationAddr.String())
	}
}

// checkTotalAssets asserts if assets on transaction equals assets on ledger
func checkTotalAssets(essence *iotago.TransactionEssence, lastTotalOnChainAssets *iscp.Assets) {
	//var chainOutput *ledgerstate.AliasOutput
	//for _, o := range essence.Outputs() {
	//	if out, ok := o.(*ledgerstate.AliasOutput); ok {
	//		chainOutput = out
	//	}
	//}
	//if chainOutput == nil {
	//	return xerrors.New("inconsistency: chain output not found")
	//}
	//balancesOnOutput := colored.BalancesFromL1Balances(chainOutput.Balances())
	//diffAssets := balancesOnOutput.Diff(lastTotalOnChainAssets)
	//// we expect assets in the chain output and total assets on-chain differs only in the amount of
	//// anti-dust tokens locked in the output. Otherwise it is inconsistency
	//if len(diffAssets) != 1 || diffAssets[colored.IOTA] != int64(ledgerstate.DustThresholdAliasOutputIOTA) {
	//	return xerrors.Errorf("inconsistency between L1 and L2 ledgers. Diff: %+v", diffAssets)
	//}
}
