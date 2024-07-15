package vmimpl

import (
	"errors"
	"math"

	"github.com/iotaledger/hive.go/lo"
	"github.com/iotaledger/hive.go/logger"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
	"github.com/iotaledger/wasp/packages/vm/core/evm/evmimpl"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/util/panicutil"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/vm/vmexceptions"
	"github.com/iotaledger/wasp/packages/vm/vmtxbuilder"
)

func Run(task *vm.VMTask) (res *vm.VMTaskResult, err error) {
	// top exception catcher for all panics
	// The VM session will be abandoned peacefully
	err = panicutil.CatchAllButDBError(func() {
		res = runTask(task)
	}, task.Log)
	if err != nil {
		task.Log.Warnf("GENERAL VM EXCEPTION: the task has been abandoned due to: %s", err.Error())
	}
	return res, err
}

// runTask runs batch of requests on VM
func runTask(task *vm.VMTask) *vm.VMTaskResult {
	if len(task.Requests) == 0 {
		panic("invalid params: must be at least 1 request")
	}

	prevL1Commitment, err := transaction.L1CommitmentFromAliasOutput(task.AnchorOutput)
	if err != nil {
		panic(err)
	}

	stateDraft, err := task.Store.NewStateDraft(task.TimeAssumption, prevL1Commitment)
	if err != nil {
		panic(err)
	}

	vmctx := &vmContext{
		task:       task,
		stateDraft: stateDraft,
	}

	vmctx.init()

	// run the batch of requests
	requestResults, numSuccess, numOffLedger, unprocessable := vmctx.runRequests(
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

	vmctx.task.Log.Debugf("runTask, ran %d requests. success: %d, offledger: %d",
		numProcessed, numSuccess, numOffLedger)

	blockIndex, l1Commitment, timestamp, rotationAddr := vmctx.extractBlock(
		numProcessed, numSuccess, numOffLedger,
		unprocessable,
	)

	vmctx.task.Log.Debugf("closed vmContext: block index: %d, state hash: %s timestamp: %v, rotationAddr: %v",
		blockIndex, l1Commitment, timestamp, rotationAddr)

	if rotationAddr == nil {
		// rotation does not happen
		taskResult.TransactionEssence, taskResult.InputsCommitment = vmctx.BuildTransactionEssence(l1Commitment, true)
		vmctx.task.Log.Debugf("runTask OUT. block index: %d", blockIndex)
	} else {
		// rotation happens
		taskResult.RotationAddress = rotationAddr
		taskResult.TransactionEssence = nil
		vmctx.task.Log.Debugf("runTask OUT: rotate to address %s", rotationAddr.String())
	}
	return taskResult
}

func (vmctx *vmContext) init() {
	vmctx.loadChainConfig()

	vmctx.withStateUpdate(func(chainState kv.KVStore) {
		vmctx.runMigrations(chainState, vmctx.task.Migrations)
		vmctx.schemaVersion = root.NewStateReaderFromChainState(chainState).GetSchemaVersion()

		// save the anchor tx ID of the current state
		blocklog.NewStateWriter(blocklog.Contract.StateSubrealm(chainState)).UpdateLatestBlockInfo(
			vmctx.task.AnchorOutputID,
		)

		// save the OutputID of the newly created tokens, foundries and NFTs in the previous block
		accountsState := vmctx.accountsStateWriterFromChainState(chainState)
		newNFTIDs := accountsState.
			UpdateLatestOutputID(vmctx.task.AnchorOutputID.TransactionID(), vmctx.task.AnchorOutput.StateIndex)

		if len(newNFTIDs) > 0 {
			for nftID, owner := range newNFTIDs {
				nft := accountsState.GetNFTData(nftID)
				if owner.Kind() == isc.AgentIDKindEthereumAddress {
					// emit an EVM event so that the mint is visible from the EVM block explorer
					vmctx.onBlockClose(
						vmctx.emitEVMEventL1NFTMint(nft.ID, owner.(*isc.EthereumAddressAgentID)),
					)
				}
			}
		}
	})

	vmctx.txbuilder = vmtxbuilder.NewAnchorTransactionBuilder(
		vmctx.task.AnchorOutput,
		vmctx.task.AnchorOutputID,
		vmctx.getAnchorOutputSD(),
		vmtxbuilder.AccountsContractRead{
			NativeTokenOutput:   vmctx.loadNativeTokenOutput,
			FoundryOutput:       vmctx.loadFoundry,
			NFTOutput:           vmctx.loadNFT,
			TotalFungibleTokens: vmctx.loadTotalFungibleTokens,
		},
	)
}

func (vmctx *vmContext) emitEVMEventL1NFTMint(nftID iotago.NFTID, owner *isc.EthereumAddressAgentID) blockCloseCallback {
	return func(reqIndex uint16) {
		// fake a request execution and insert a Mint event on the EVM
		reqCtx := vmctx.newRequestContext(isc.NewImpersonatedOffLedgerRequest(&isc.OffLedgerRequestData{}).WithSenderAddress(&iotago.Ed25519Address{}), reqIndex)
		reqCtx.pushCallContext(evm.Contract.Hname(), nil, nil, nil)
		ctx := NewSandbox(reqCtx)
		evmimpl.AddDummyTxWithTransferEvents(
			ctx,
			owner.EthAddress(),
			isc.NewEmptyAssets().AddNFTs(nftID),
			nil,
			false,
		)
		reqCtx.uncommittedState.Mutations().ApplyTo(vmctx.stateDraft)
	}
}

func (vmctx *vmContext) getAnchorOutputSD() uint64 {
	// get the total L2 funds in accounting
	totalL2Funds := vmctx.loadTotalFungibleTokens()
	return vmctx.task.AnchorOutput.Amount - totalL2Funds.BaseTokens
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
	requestIndexCounter := uint16(0)
	for reqIndex := 0; reqIndex < len(allReqs); reqIndex++ {
		req := allReqs[reqIndex]
		result, unprocessableToRetry, skipReason := vmctx.runRequest(req, requestIndexCounter, maintenanceMode)
		if skipReason != nil {
			if errors.Is(vmexceptions.ErrNotEnoughFundsForSD, skipReason) {
				unprocessable = append(unprocessable, req.(isc.OnLedgerRequest))
			}

			// some requests are just ignored (deterministically)
			log.Infof("request skipped (ignored) by the VM: %s, reason: %v",
				req.ID().String(), skipReason)
			continue
		}

		requestIndexCounter++
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

		// abort if num of requests is above max_uint16.
		if reqIndex+1 == math.MaxUint16 {
			log.Warnf("aborting vm run due to excessive number of requests. total: %d, executed: %d", len(reqs), reqIndex+1)
			break
		}
	}
	return results, numSuccess, numOffLedger, unprocessable
}
