package runvm

import (
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/util/panicutil"
	"github.com/iotaledger/wasp/packages/vm"
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

func runRequests(vmctx *vmcontext.VMContext, reqs []isc.Request, startRequestIndex uint16, log *logger.Logger) (results []*vm.RequestResult, numSuccess uint16, numOffLedger uint16) {
	results = []*vm.RequestResult{}
	reqIndexInTheBlock := startRequestIndex

	// main loop over the batch of requests
	for _, req := range reqs {
		result, skipReason := vmctx.RunTheRequest(req, reqIndexInTheBlock)
		if skipReason != nil {
			// some requests are just ignored (deterministically)
			log.Infof("request skipped (ignored) by the VM: %s, reason: %v",
				req.ID().String(), skipReason)
			continue
		}
		results = append(results, result)
		reqIndexInTheBlock++
		if req.IsOffLedger() {
			numOffLedger++
		}

		if result.Receipt.Error == nil {
			numSuccess++
		} else {
			log.Debugf("runTask, ERROR running request: %s, error: %v", req.ID().String(), result.Receipt.Error)
		}
	}
	return results, numSuccess, numOffLedger
}

// runTask runs batch of requests on VM
func runTask(task *vm.VMTask) {
	vmctx := vmcontext.CreateVMContext(task)

	vmctx.OpenBlockContexts()

	// run the batch of requests
	results, numSuccess, numOffLedger := runRequests(vmctx, task.Requests, 0, task.Log)
	{
		// run any scheduled retry of "unprocessable" requests
		results2, numSuccess2, numOffLedger2 := runRequests(vmctx, task.UnprocessableToRetry, uint16(len(results)), task.Log)
		vmctx.RemoveUnprocessable(results2)
		if numOffLedger2 != 0 {
			panic("offledger request executed as 'unprocessable retry', this cannot happen")
		}
		task.Results = results
		task.Results = append(task.Results, results2...)
		numSuccess += numSuccess2
	}

	vmctx.AssertConsistentGasTotals()

	if !task.WillProduceBlock() {
		return
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
		task.ResultTransactionEssence, task.ResultInputsCommitment = vmctx.BuildTransactionEssence(l1Commitment, true)
		task.Log.Debugf("runTask OUT. block index: %d", blockIndex)
	} else {
		// rotation happens
		task.RotationAddress = rotationAddr
		task.ResultTransactionEssence = nil
		task.Log.Debugf("runTask OUT: rotate to address %s", rotationAddr.String())
	}
}
