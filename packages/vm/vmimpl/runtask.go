package vmimpl

import (
	"errors"
	"math"

	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/hive.go/logger"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/vmexceptions"
)

// runTask runs batch of requests on VM
func (vmctx *vmContext) run() {
	maintenanceMode := governance.NewStateAccess(vmctx.taskResult.StateDraft).MaintenanceStatus()

	vmctx.openBlockContexts()

	// run the batch of requests
	results, numSuccess, numOffLedger, unprocessable := vmctx.runRequests(
		vmctx.task.Requests,
		maintenanceMode,
		vmctx.task.Log,
	)
	vmctx.taskResult.RequestResults = results

	vmctx.assertConsistentGasTotals()

	if !vmctx.task.WillProduceBlock() {
		return
	}

	numProcessed := uint16(len(vmctx.taskResult.RequestResults))

	vmctx.task.Log.Debugf("runTask, ran %d requests. success: %d, offledger: %d",
		numProcessed, numSuccess, numOffLedger)

	blockIndex, l1Commitment, timestamp, rotationAddr := vmctx.extractBlock(
		numProcessed, numSuccess, numOffLedger,
		unprocessable,
	)

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
	maintenanceMode bool,
	log *logger.Logger,
) (
	results []*vm.RequestResult,
	numSuccess uint16,
	numOffLedger uint16,
	unprocessable []isc.OnLedgerRequest,
) {
	results = []*vm.RequestResult{}
	allReqs := lo.CopySlice(reqs)

	// main loop over the batch of requests
	for reqIndex := 0; reqIndex < len(allReqs); reqIndex++ {
		req := allReqs[reqIndex]
		result, unprocessableToRetry, skipReason := vmctx.runRequest(req, uint16(reqIndex), maintenanceMode)
		if skipReason != nil {
			if errors.Is(vmexceptions.ErrNotEnoughFundsForSD, skipReason) {
				unprocessable = append(unprocessable, req.(isc.OnLedgerRequest))
			}

			// some requests are just ignored (deterministically)
			log.Infof("request skipped (ignored) by the VM: %s, reason: %v",
				req.ID().String(), skipReason)
			continue
		}
		results = append(results, result)
		if req.IsOffLedger() {
			numOffLedger++
		}

		if result.Receipt.Error != nil {
			log.Debugf("runTask, ERROR running request: %s, error: %v", req.ID().String(), result.Receipt.Error)
			continue
		}
		numSuccess++

		isRetry := reqIndex >= len(reqs)
		if isRetry {
			vmctx.removeUnprocessable(req.ID())
		}
		for _, retry := range unprocessableToRetry {
			if len(allReqs) >= math.MaxUint16 {
				log.Warnf("cannot process request to be retried %s (retry requested in %s): too many requests in block",
					retry.ID(), req.ID())
			} else {
				allReqs = append(allReqs, retry)
			}
		}
	}
	return results, numSuccess, numOffLedger, unprocessable
}
