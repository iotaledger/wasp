package vmimpl

import (
	"github.com/iotaledger/hive.go/logger"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/vm"
)

// runTask runs batch of requests on VM
func (vmctx *vmContext) run() {
	vmctx.openBlockContexts()

	// run the batch of requests
	results, numSuccess, numOffLedger := vmctx.runRequests(vmctx.task.Requests, 0, vmctx.task.Log)
	{
		// run any scheduled retry of "unprocessable" requests
		results2, numSuccess2, numOffLedger2 := vmctx.runRequests(vmctx.task.UnprocessableToRetry, uint16(len(results)), vmctx.task.Log)
		vmctx.removeUnprocessable(results2)
		if numOffLedger2 != 0 {
			panic("offledger request executed as 'unprocessable retry', this cannot happen")
		}
		vmctx.taskResult.RequestResults = results
		vmctx.taskResult.RequestResults = append(vmctx.taskResult.RequestResults, results2...)
		numSuccess += numSuccess2
	}

	vmctx.assertConsistentGasTotals()

	if !vmctx.task.WillProduceBlock() {
		return
	}

	numProcessed := uint16(len(vmctx.taskResult.RequestResults))

	vmctx.task.Log.Debugf("runTask, ran %d requests. success: %d, offledger: %d",
		numProcessed, numSuccess, numOffLedger)

	blockIndex, l1Commitment, timestamp, rotationAddr := vmctx.extractBlock(
		numProcessed, numSuccess, numOffLedger)

	vmctx.task.Log.Debugf("closed VMContext: block index: %d, state hash: %s timestamp: %v, rotationAddr: %v",
		blockIndex, l1Commitment, timestamp, rotationAddr)

	if rotationAddr == nil {
		// rotation does not happen
		vmctx.taskResult.TransactionEssence, vmctx.taskResult.InputsCommitment = vmctx.BuildTransactionEssence(l1Commitment, true)
		vmctx.task.Log.Debugf("runTask OUT. block index: %d", blockIndex)
	} else {
		// rotation happens
		vmctx.taskResult.RotationAddress = rotationAddr
		vmctx.taskResult.TransactionEssence = nil
		vmctx.task.Log.Debugf("runTask OUT: rotate to address %s", rotationAddr.String())
	}
}

func (vmctx *vmContext) runRequests(
	reqs []isc.Request,
	startRequestIndex uint16,
	log *logger.Logger,
) (
	results []*vm.RequestResult,
	numSuccess uint16,
	numOffLedger uint16,
) {
	results = []*vm.RequestResult{}
	reqIndexInTheBlock := startRequestIndex

	// main loop over the batch of requests
	for _, req := range reqs {
		result, skipReason := vmctx.runRequest(req, reqIndexInTheBlock)
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
