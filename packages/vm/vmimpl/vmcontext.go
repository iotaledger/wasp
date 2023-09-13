package vmimpl

import (
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/buffered"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"
	"github.com/iotaledger/wasp/packages/vm/core/blob"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
	"github.com/iotaledger/wasp/packages/vm/core/evm/evmimpl"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/execution"
	"github.com/iotaledger/wasp/packages/vm/gas"
	"github.com/iotaledger/wasp/packages/vm/vmtxbuilder"
)

// vmContext represents state of the chain during one run of the VM while processing
// a batch of requests. vmContext object mutates with each request in the batch.
// The vmContext is created from immutable vm.VMTask object and UTXO state of the
// chain address contained in the statetxbuilder.Builder
type vmContext struct {
	task *vm.VMTask

	stateDraft state.StateDraft
	txbuilder  *vmtxbuilder.AnchorTransactionBuilder
	chainInfo  *isc.ChainInfo
	blockGas   blockGas
}

type blockGas struct {
	burned     uint64
	feeCharged uint64
}

type requestContext struct {
	vm *vmContext

	uncommittedState  *buffered.BufferedKVStore
	callStack         []*callContext
	req               isc.Request
	numPostedOutputs  int
	requestIndex      uint16
	requestEventIndex uint16
	entropy           hashing.HashValue
	onWriteReceipt    []coreCallbackFunc
	gas               requestGas
	// SD charged to consume the current request
	sdCharged uint64
	// requests that the sender asked to retry
	unprocessableToRetry []isc.OnLedgerRequest
	// snapshots taken via ctx.TakeStateSnapshot()
	snapshots []stateSnapshot
}

type stateSnapshot struct {
	txb   *vmtxbuilder.AnchorTransactionBuilder
	state *buffered.BufferedKVStore
}

type requestGas struct {
	// is gas burn enabled
	burnEnabled bool
	// max tokens that can be charged for gas fee
	maxTokensToSpendForGasFee uint64
	// final gas budget set for the run
	budgetAdjusted uint64
	// gas already burned
	burned uint64
	// tokens charged
	feeCharged uint64
	// burn history. If disabled, it is nil
	burnLog *gas.BurnLog
}

type coreCallbackFunc struct {
	contract isc.Hname
	callback isc.CoreCallbackFunc
}

var _ execution.WaspCallContext = &requestContext{}

type callContext struct {
	caller   isc.AgentID // calling agent
	contract isc.Hname   // called contract
	params   isc.Params  // params passed
	// MUTABLE: allowance budget left after TransferAllowedFunds
	// TODO: should be in requestContext?
	allowanceAvailable *isc.Assets
}

func (vmctx *vmContext) withStateUpdate(f func(chainState kv.KVStore)) {
	chainState := buffered.NewBufferedKVStore(vmctx.stateDraft)
	f(chainState)
	chainState.Mutations().ApplyTo(vmctx.stateDraft)
}

// extractBlock does the closing actions on the block
// return nil for normal block and rotation address for rotation block
func (vmctx *vmContext) extractBlock(
	numRequests, numSuccess, numOffLedger uint16,
	unprocessable []isc.OnLedgerRequest,
) (uint32, *state.L1Commitment, time.Time, iotago.Address) {
	var rotationAddr iotago.Address
	vmctx.withStateUpdate(func(chainState kv.KVStore) {
		rotationAddr = vmctx.saveBlockInfo(numRequests, numSuccess, numOffLedger)
		withContractState(chainState, evm.Contract, func(s kv.KVStore) {
			evmimpl.MintBlock(s, vmctx.chainInfo, vmctx.task.TimeAssumption)
		})
		vmctx.saveInternalUTXOs(unprocessable)
	})

	block := vmctx.task.Store.ExtractBlock(vmctx.stateDraft)

	l1Commitment := block.L1Commitment()

	blockIndex := vmctx.stateDraft.BlockIndex()
	timestamp := vmctx.stateDraft.Timestamp()

	return blockIndex, l1Commitment, timestamp, rotationAddr
}

func (vmctx *vmContext) checkRotationAddress() (ret iotago.Address) {
	withContractState(vmctx.stateDraft, governance.Contract, func(s kv.KVStore) {
		ret = governance.GetRotationAddress(s)
	})
	return
}

// saveBlockInfo is in the blocklog partition context. Returns rotation address if this block is a rotation block
func (vmctx *vmContext) saveBlockInfo(numRequests, numSuccess, numOffLedger uint16) iotago.Address {
	if rotationAddress := vmctx.checkRotationAddress(); rotationAddress != nil {
		// block was marked fake by the governance contract because it is a committee rotation.
		// There was only on request in the block
		// We skip saving block information in order to avoid inconsistencies
		return rotationAddress
	}

	blockInfo := &blocklog.BlockInfo{
		SchemaVersion:         blocklog.BlockInfoLatestSchemaVersion,
		Timestamp:             vmctx.stateDraft.Timestamp(),
		TotalRequests:         numRequests,
		NumSuccessfulRequests: numSuccess,
		NumOffLedgerRequests:  numOffLedger,
		PreviousAliasOutput:   isc.NewAliasOutputWithID(vmctx.task.AnchorOutput, vmctx.task.AnchorOutputID),
		GasBurned:             vmctx.blockGas.burned,
		GasFeeCharged:         vmctx.blockGas.feeCharged,
	}

	withContractState(vmctx.stateDraft, blocklog.Contract, func(s kv.KVStore) {
		blocklog.SaveNextBlockInfo(s, blockInfo)
		blocklog.Prune(s, blockInfo.BlockIndex(), vmctx.chainInfo.BlockKeepAmount)
	})
	vmctx.task.Log.Debugf("saved blockinfo:\n%s", blockInfo)
	return nil
}

// saveInternalUTXOs relies on the order of the outputs in the anchor tx. If that order changes, this will be broken.
// Anchor Transaction outputs order must be:
// 0. Anchor Output
// 1. NativeTokens
// 2. Foundries
// 3. NFTs
// 4. produced outputs
// 5. unprocessable requests
func (vmctx *vmContext) saveInternalUTXOs(unprocessable []isc.OnLedgerRequest) {
	// create a mock AO, with a nil statecommitment, just to calculate changes in the minimum SD
	mockAO := vmctx.txbuilder.CreateAnchorOutput(vmctx.stateMetadata(state.L1CommitmentNil))
	newMinSD := parameters.L1().Protocol.RentStructure.MinRent(mockAO)
	oldMinSD := vmctx.txbuilder.AnchorOutputStorageDeposit()
	changeInSD := int64(oldMinSD) - int64(newMinSD)

	if changeInSD != 0 {
		vmctx.task.Log.Debugf("adjusting commonAccount because AO SD cost changed, old:%d new:%d", oldMinSD, newMinSD)
		// update the commonAccount with the change in SD cost
		withContractState(vmctx.stateDraft, accounts.Contract, func(s kv.KVStore) {
			accounts.AdjustAccountBaseTokens(s, accounts.CommonAccount(), changeInSD, vmctx.ChainID())
		})
	}

	nativeTokenIDsToBeUpdated, nativeTokensToBeRemoved := vmctx.txbuilder.NativeTokenRecordsToBeUpdated()
	// IMPORTANT: do not iterate by this map, order of the slice above must be respected
	nativeTokensMap := vmctx.txbuilder.NativeTokenOutputsByTokenIDs(nativeTokenIDsToBeUpdated)

	foundryIDsToBeUpdated, foundriesToBeRemoved := vmctx.txbuilder.FoundriesToBeUpdated()
	// IMPORTANT: do not iterate by this map, order of the slice above must be respected
	foundryOutputsMap := vmctx.txbuilder.FoundryOutputsBySN(foundryIDsToBeUpdated)

	NFTOutputsToBeAdded, NFTOutputsToBeRemoved, MintedNFTOutputs := vmctx.txbuilder.NFTOutputsToBeUpdated()

	outputIndex := uint16(1)

	withContractState(vmctx.stateDraft, accounts.Contract, func(s kv.KVStore) {
		// update native token outputs
		for _, ntID := range nativeTokenIDsToBeUpdated {
			vmctx.task.Log.Debugf("saving NT %s, outputIndex: %d", ntID, outputIndex)
			accounts.SaveNativeTokenOutput(s, nativeTokensMap[ntID], outputIndex)
			outputIndex++
		}
		for _, id := range nativeTokensToBeRemoved {
			vmctx.task.Log.Debugf("deleting NT %s", id)
			accounts.DeleteNativeTokenOutput(s, id)
		}

		// update foundry UTXOs
		for _, foundryID := range foundryIDsToBeUpdated {
			vmctx.task.Log.Debugf("saving foundry %d, outputIndex: %d", foundryID, outputIndex)
			accounts.SaveFoundryOutput(s, foundryOutputsMap[foundryID], outputIndex)
			outputIndex++
		}
		for _, sn := range foundriesToBeRemoved {
			vmctx.task.Log.Debugf("deleting foundry %d", sn)
			accounts.DeleteFoundryOutput(s, sn)
		}

		// update NFT Outputs
		for _, out := range NFTOutputsToBeAdded {
			vmctx.task.Log.Debugf("saving NFT %s, outputIndex: %d", out.NFTID, outputIndex)
			accounts.SaveNFTOutput(s, out, outputIndex)
			outputIndex++
		}
		for _, out := range NFTOutputsToBeRemoved {
			vmctx.task.Log.Debugf("deleting NFT %s", out.NFTID)
			accounts.DeleteNFTOutput(s, out.NFTID)
		}

		for positionInMintedList := range MintedNFTOutputs {
			vmctx.task.Log.Debugf("minted NFT on output index: %d", outputIndex)
			accounts.SaveMintedNFTOutput(s, uint16(positionInMintedList), outputIndex)
			outputIndex++
		}
	})

	// add unprocessable requests
	vmctx.storeUnprocessable(vmctx.stateDraft, unprocessable, outputIndex)
}

func (vmctx *vmContext) removeUnprocessable(reqID isc.RequestID) {
	withContractState(vmctx.stateDraft, blocklog.Contract, func(s kv.KVStore) {
		blocklog.RemoveUnprocessable(s, reqID)
	})
}

func (vmctx *vmContext) assertConsistentGasTotals(requestResults []*vm.RequestResult) {
	var sumGasBurned, sumGasFeeCharged uint64

	for _, r := range requestResults {
		sumGasBurned += r.Receipt.GasBurned
		sumGasFeeCharged += r.Receipt.GasFeeCharged
	}
	if vmctx.blockGas.burned != sumGasBurned {
		panic("vmctx.gasBurnedTotal != sumGasBurned")
	}
	if vmctx.blockGas.feeCharged != sumGasFeeCharged {
		panic("vmctx.gasFeeChargedTotal != sumGasFeeCharged")
	}
}

func (vmctx *vmContext) locateProgram(chainState kv.KVStore, programHash hashing.HashValue) (vmtype string, binary []byte, err error) {
	withContractState(chainState, blob.Contract, func(s kv.KVStore) {
		vmtype, binary, err = blob.LocateProgram(s, programHash)
	})
	return vmtype, binary, err
}
