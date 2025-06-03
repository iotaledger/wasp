// Package vmimpl is the implementation of a wasp vm
package vmimpl

import (
	"time"

	"github.com/iotaledger/wasp/packages/coin"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/buffered"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/core/evm"
	"github.com/iotaledger/wasp/packages/vm/core/evm/evmimpl"
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
	txbuilder  vmtxbuilder.TransactionBuilder
	chainInfo  *isc.ChainInfo
	blockGas   blockGas

	onBlockCloseCallbacks []blockCloseCallback

	schemaVersion isc.SchemaVersion
}
type blockCloseCallback func(requestIndex uint16)

type blockGas struct {
	burned     uint64
	feeCharged coin.Value
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
	// snapshots taken via ctx.TakeStateSnapshot()
	snapshots []stateSnapshot
}

type stateSnapshot struct {
	txb   vmtxbuilder.TransactionBuilder
	state *buffered.BufferedKVStore
}

type requestGas struct {
	// is gas burn enabled
	burnEnabled bool
	// max tokens that can be charged for gas fee
	maxTokensToSpendForGasFee coin.Value
	// final gas budget set for the run
	budgetAdjusted uint64
	// gas already burned
	burned uint64
	// tokens charged
	feeCharged coin.Value
	// burn history. If disabled, it is nil
	burnLog *gas.BurnLog
	// used to allow tracing stardust requests
	enforceGasBurned *vm.EnforceGasBurned
}

type coreCallbackFunc struct {
	contract isc.Hname
	callback isc.CoreCallbackFunc
}

var _ execution.WaspCallContext = &requestContext{}

type callContext struct {
	caller   isc.AgentID       // calling agent
	contract isc.Hname         // called contract
	params   isc.CallArguments // params passed
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
func (vmctx *vmContext) extractBlock(
	numRequests, numSuccess, numOffLedger uint16,
) (uint32, *state.L1Commitment, time.Time) {
	vmctx.withStateUpdate(func(chainState kv.KVStore) {
		vmctx.saveBlockInfo(chainState, numRequests, numSuccess, numOffLedger)
		evmimpl.MintBlock(evm.Contract.StateSubrealm(chainState), vmctx.chainInfo, vmctx.task.Timestamp)
	})

	block := vmctx.task.Store.ExtractBlock(vmctx.stateDraft)

	l1Commitment := block.L1Commitment()

	blockIndex := vmctx.stateDraft.BlockIndex()
	timestamp := vmctx.stateDraft.Timestamp()

	return blockIndex, l1Commitment, timestamp
}

// saveBlockInfo is in the blocklog partition context
func (vmctx *vmContext) saveBlockInfo(chainState kv.KVStore, numRequests, numSuccess, numOffLedger uint16) {
	blockInfo := &blocklog.BlockInfo{
		SchemaVersion:         blocklog.BlockInfoLatestSchemaVersion,
		BlockIndex:            vmctx.stateDraft.BlockIndex(),
		Timestamp:             vmctx.stateDraft.Timestamp(),
		PreviousAnchor:        vmctx.task.Anchor,
		L1Params:              vmctx.task.L1Params,
		TotalRequests:         numRequests,
		NumSuccessfulRequests: numSuccess,
		NumOffLedgerRequests:  numOffLedger,
		GasBurned:             vmctx.blockGas.burned,
		GasFeeCharged:         vmctx.blockGas.feeCharged,
		Entropy:               vmctx.task.Entropy,
	}

	blocklogState := blocklog.NewStateWriter(blocklog.Contract.StateSubrealm(chainState))
	blocklogState.SaveNextBlockInfo(blockInfo)
	blocklogState.Prune(blockInfo.BlockIndex, vmctx.chainInfo.BlockKeepAmount)
	vmctx.task.Log.LogDebugf("saved blockinfo:\n%s", blockInfo)
}

func (vmctx *vmContext) assertConsistentGasTotals(requestResults []*vm.RequestResult) {
	var sumGasBurned uint64
	var sumGasFeeCharged coin.Value

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
