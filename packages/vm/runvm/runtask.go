package runvm

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/util"

	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/vmcontext"
)

type VMRunner struct{}

func (r VMRunner) Run(task *vm.VMTask) {
	// optimistic read panic catcher for the whole VM task
	err := util.CatchPanicReturnError(
		func() { runTask(task) },
		coreutil.ErrorStateInvalidated,
	)
	if err != nil {
		task.Log.Warnf("VM task has been abandoned due to invalidated state. ACS session id: %d", task.ACSSessionID)
	}
}

func NewVMRunner() VMRunner {
	return VMRunner{}
}

// runTask runs batch of requests on VM
func runTask(task *vm.VMTask) {
	vmctx := vmcontext.CreateVMContext(task)

	var lastResult dict.Dict
	var lastErr error

	var numOffLedger, numSuccess uint16
	reqIndexInTheBlock := 0

	// main loop over the batch of requests
	for _, req := range task.Requests {
		if skipReason := vmctx.RunTheRequest(req, uint16(reqIndexInTheBlock)); skipReason != nil {
			// some requests are just ignored (deterministically)
			task.Log.Warnf("request skipped (ignored) by the VM: %s, reason: %v",
				req.ID().String(), skipReason)
			continue
		}
		reqIndexInTheBlock++
		if req.IsOffLedger() {
			numOffLedger++
		}
		// get the last result from the call to the entry point. It is used by Solo only
		lastResult, lastErr = vmctx.GetResult()

		task.ProcessedRequestsCount++
		if lastErr == nil {
			numSuccess++
		} else {
			task.Log.Debugf("runTask, ERROR running request: %s, error: %v", req.ID().String(), lastErr)
		}
	}

	task.Log.Debugf("runTask, ran %d requests. success: %d, offledger: %d",
		task.ProcessedRequestsCount, numSuccess, numOffLedger)

	if task.ProcessedRequestsCount == 0 {
		// empty result. Abandon without closing
		task.OnFinish(nil, nil, nil)
	}

	blockIndex, stateCommitment, timestamp, rotationAddr := vmctx.CloseVMContext(
		task.ProcessedRequestsCount, numSuccess, numOffLedger)

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
	// call callback closure
	task.OnFinish(lastResult, lastErr, nil)
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
