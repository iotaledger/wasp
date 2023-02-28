package runvm

import (
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/subrealm"
	"github.com/iotaledger/wasp/packages/util/panicutil"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/vmcontext"
)

type VMRunner struct{}

func (r VMRunner) Run(task *vm.VMTask) error {
	// top exception catcher for all panics
	// The VM session will be abandoned peacefully
	err := panicutil.CatchAllButDBError(func() {
		runTask(task)
	}, task.Log)
	if err != nil {
		task.Log.Warnf("GENERAL VM EXCEPTION: the task has been abandoned due to: %s", err.Error())
	}
	return err
}

func NewVMRunner() vm.VMRunner {
	return VMRunner{}
}

// runTask runs batch of requests on VM
func runTask(task *vm.VMTask) {
	vmctx := vmcontext.CreateVMContext(task)

	var numOffLedger, numSuccess uint16
	reqIndexInTheBlock := 0

	vmctx.OpenBlockContexts()

	// main loop over the batch of requests
	for _, req := range task.Requests {
		result, skipReason := vmctx.RunTheRequest(req, uint16(reqIndexInTheBlock))
		if skipReason != nil {
			// some requests are just ignored (deterministically)
			task.Log.Infof("request skipped (ignored) by the VM: %s, reason: %v",
				req.ID().String(), skipReason)
			continue
		}
		task.Results = append(task.Results, result)
		reqIndexInTheBlock++
		if req.IsOffLedger() {
			numOffLedger++
		}

		if result.Receipt.Error == nil {
			numSuccess++
		} else {
			task.Log.Debugf("runTask, ERROR running request: %s, error: %v", req.ID().String(), result.Receipt.Error)
		}
		vmctx.AssertConsistentGasTotals()
	}

	// closing the task and producing a new block is not needed if we are estimating gas
	if task.EstimateGasMode {
		return
	}

	{
		accountsState := subrealm.NewReadOnly(task.StateDraft, kv.Key(accounts.Contract.Hname().Bytes()))
		accounts.CheckLedger(accountsState, "runTask")
	}

	numProcessed := uint16(len(task.Results))

	task.Log.Debugf("runTask, ran %d requests. success: %d, offledger: %d",
		numProcessed, numSuccess, numOffLedger)

	blockIndex, l1Commitment, timestamp, rotationAddr := vmctx.CloseVMContext(
		numProcessed, numSuccess, numOffLedger)

	task.Log.Debugf("closed VMContext: block index: %d, state hash: %s timestamp: %v, rotationAddr: %v",
		blockIndex, l1Commitment, timestamp, rotationAddr)

	if rotationAddr == nil {
		// rotation does not happen
		task.ResultTransactionEssence, task.ResultInputsCommitment = vmctx.BuildTransactionEssence(l1Commitment)

		// TODO extract latest total assets
		checkTotalAssets(task.ResultTransactionEssence, nil)

		task.Log.Debugf("runTask OUT. block index: %d, %s", blockIndex, l1Commitment.String())
	} else {
		// rotation happens
		task.RotationAddress = rotationAddr
		task.ResultTransactionEssence = nil
		task.Log.Debugf("runTask OUT: rotate to address %s", rotationAddr.String())
	}
}

// checkTotalAssets asserts if assets on the L1 transaction equals assets on the chain's ledger
func checkTotalAssets(_ *iotago.TransactionEssence, _ *isc.Assets) {
	// TODO implement
}
