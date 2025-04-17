package vmimpl

import (
	"math"

	"github.com/samber/lo"

	"github.com/iotaledger/hive.go/log"

	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/util/panicutil"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/vm/vmtxbuilder"
)

func Run(task *vm.VMTask) (res *vm.VMTaskResult, err error) {
	// top exception catcher for all panics
	// The VM session will be abandoned peacefully
	err = panicutil.CatchAllButDBError(func() {
		res = runTask(task)
	}, task.Log)
	if err != nil {
		task.Log.LogWarnf("GENERAL VM EXCEPTION: the task has been abandoned due to: %s", err.Error())
	}
	return res, err
}

func newVMContext(
	task *vm.VMTask,
	stateDraft state.StateDraft,
	txbuilder vmtxbuilder.TransactionBuilder,
) *vmContext {
	return &vmContext{
		task:       task,
		stateDraft: stateDraft,
		txbuilder:  txbuilder,
	}
}

func (vmctx *vmContext) calculateTopUpFee() coin.Value {
	targetValue := governance.NewStateReaderFromChainState(vmctx.stateDraft).GetGasCoinTargetValue()
	gasCoinBalance := vmctx.task.GasCoin.Value

	topUp := coin.Value(0)
	if gasCoinBalance < targetValue {
		topUp = targetValue - gasCoinBalance
	}

	bal := vmctx.commonAccountBalance()
	if bal < topUp {
		vmctx.task.Log.LogWarnf(
			"not enough tokens in common account for topping up gas coin (has %d, want %d)",
			bal,
			topUp,
		)
		topUp = bal
	}

	vmctx.task.Log.LogDebugf(
		"calculateTopUpFee: gasCoinBalance: %d, target: %d, commonAccountBalance: %d, topUp: %d",
		gasCoinBalance,
		targetValue,
		bal,
		topUp,
	)
	return topUp
}

// runTask runs batch of requests on VM
func runTask(task *vm.VMTask) *vm.VMTaskResult {
	if len(task.Requests) == 0 {
		panic("invalid params: must be at least 1 request")
	}

	prevL1Commitment := lo.Must(transaction.L1CommitmentFromAnchor(task.Anchor))
	stateDraft := lo.Must(task.Store.NewStateDraft(task.Timestamp, prevL1Commitment))
	txbuilder := vmtxbuilder.NewAnchorTransactionBuilder(task.Anchor.ISCPackage(), task.Anchor, task.Anchor.Owner())
	vmctx := newVMContext(task, stateDraft, txbuilder)
	vmctx.init()

	// run the batch of requests
	requestResults, numSuccess, numOffLedger := vmctx.runRequests(
		vmctx.task.Requests,
		governance.NewStateReaderFromChainState(stateDraft).GetMaintenanceStatus(),
		vmctx.task.Log,
	)
	numProcessed := uint16(len(requestResults))

	// execute onBlockClose callbacks
	for _, callback := range vmctx.onBlockCloseCallbacks {
		callback(numProcessed + 1)
	}

	vmctx.assertConsistentGasTotals(requestResults)

	taskResult := &vm.VMTaskResult{
		Task:           task,
		StateDraft:     stateDraft,
		RequestResults: requestResults,
	}

	if !vmctx.task.WillProduceBlock() {
		return taskResult
	}

	vmctx.task.Log.LogDebugf("runTask, ran %d requests. success: %d, offledger: %d",
		numProcessed, numSuccess, numOffLedger)

	topUpFee := vmctx.calculateTopUpFee()
	if topUpFee > 0 {
		vmctx.deductTopUpFeeFromCommonAccount(topUpFee)
	}

	blockIndex, l1Commitment, timestamp := vmctx.extractBlock(
		numProcessed, numSuccess, numOffLedger,
	)

	vmctx.task.Log.LogDebugf("closed vmContext: block index: %d, state hash: %s timestamp: %v",
		blockIndex, l1Commitment, timestamp)

	taskResult.StateMetadata = vmctx.StateMetadata(l1Commitment, task.GasCoin)
	vmctx.task.Log.LogDebugf("runTask OUT. block index: %d", blockIndex)

	taskResult.UnsignedTransaction = vmctx.txbuilder.BuildTransactionEssence(
		taskResult.StateMetadata,
		uint64(topUpFee),
	)
	return taskResult
}

func (vmctx *vmContext) init() {
	vmctx.loadChainConfig()

	vmctx.withStateUpdate(func(chainState kv.KVStore) {
		vmctx.runMigrations(chainState, vmctx.task.Migrations)
		vmctx.schemaVersion = root.NewStateReaderFromChainState(chainState).GetSchemaVersion()

		// TODO
		// save the ObjectID of the newly created tokens, foundries and objects in the previous block
		/*
			accountsState := vmctx.accountsStateWriterFromChainState(chainState)
			newNFTIDs := accountsState.
				UpdateLatestOutputID(vmctx.task.AnchorOutputID, vmctx.task.AnchorOutput.StateIndex)
		*/
	})
}

func (vmctx *vmContext) runRequests(
	reqs []isc.Request,
	maintenanceMode bool,
	log log.Logger,
) (
	results []*vm.RequestResult,
	numSuccess uint16,
	numOffLedger uint16,
) {
	results = []*vm.RequestResult{}

	// main loop over the batch of requests
	requestIndexCounter := uint16(0)
	for reqIndex := 0; reqIndex < len(reqs); reqIndex++ {
		req := reqs[reqIndex]
		result, skipReason := vmctx.runRequest(req, requestIndexCounter, maintenanceMode)
		if skipReason != nil {
			// some requests are just ignored (deterministically)
			log.LogInfof("request skipped (ignored) by the VM: %s, reason: %v",
				req.ID().String(), skipReason)
			continue
		}

		requestIndexCounter++
		results = append(results, result)
		if req.IsOffLedger() {
			numOffLedger++
		}

		if result.Receipt.Error != nil {
			log.LogDebugf("runTask, ERROR running request: %s, error: %v", req.ID().String(), result.Receipt.Error)
			continue
		}
		numSuccess++

		// abort if num of requests is above max_uint16.
		if reqIndex+1 == math.MaxUint16 {
			log.LogWarnf("aborting vm run due to excessive number of requests. total: %d, executed: %d", len(reqs), reqIndex+1)
			break
		}
	}
	return results, numSuccess, numOffLedger
}
