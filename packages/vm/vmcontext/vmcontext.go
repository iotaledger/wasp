package vmcontext

import (
	"time"

	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/vm/core/accounts"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/iotaledger/wasp/packages/vm/core/governance"
	"github.com/iotaledger/wasp/packages/vm/core/root"
	"github.com/iotaledger/wasp/packages/vm/vmcontext/vmtxbuilder"
	"golang.org/x/xerrors"
)

// VMContext represents state of the chain during one run of the VM while processing
// a batch of requests. VMContext object mutates with each request in the bathc.
// The VMContext is created from immutable vm.VMTask object and UTXO state of the
// chain address contained in the statetxbuilder.Builder
type VMContext struct {
	task *vm.VMTask
	// same for the block
	chainOwnerID         *iscp.AgentID
	virtualState         state.VirtualStateAccess
	finalStateTimestamp  time.Time
	blockContext         map[iscp.Hname]*blockContext
	blockContextCloseSeq []iscp.Hname
	blockOutputCount     uint8
	txbuilder            *vmtxbuilder.AnchorTransactionBuilder
	// ---- request context
	chainInfo          *governance.ChainInfo
	req                iscp.RequestData
	requestIndex       uint16
	requestEventIndex  uint16
	currentStateUpdate state.StateUpdate
	entropy            hashing.HashValue
	contractRecord     *root.ContractRecord
	lastError          error
	lastResult         dict.Dict
	callStack          []*callContext
	// --- gas related
	// gas from the request
	gasBudgetFromRequest uint64
	// max gas budget capped by the number of tokens in the sender's account
	gasBudgetAffordable uint64
	// final gas budget set for the run
	gasBudget uint64
	// gas already burned
	gasBurned uint64
}

type callContext struct {
	caller   *iscp.AgentID // calling agent
	contract iscp.Hname    // called contract
	params   dict.Dict     // params passed
	transfer *iscp.Assets  // transfer passed
}

type blockContext struct {
	obj     interface{}
	onClose func(interface{})
}

// CreateVMContext creates a context for the whole batch run
func CreateVMContext(task *vm.VMTask) *VMContext {
	// assert consistency. It is a bit redundant double check
	if len(task.Requests) == 0 {
		// should never happen
		panic(xerrors.Errorf("CreateVMContext.invalid params: must be at least 1 request"))
	}
	stateData, err := iscp.StateDataFromBytes(task.AnchorOutput.StateMetadata)
	if err != nil {
		// should never happen
		panic(xerrors.Errorf("CreateVMContext: can't parse state data from chain input %w", err))
	}
	// we create optimistic state access wrapper to be used inside the VM call.
	// It will panic any time the state is accessed.
	// The panic will be caught above and VM call will be abandoned peacefully
	optimisticStateAccess := state.WrapMustOptimisticVirtualStateAccess(task.VirtualStateAccess, task.SolidStateBaseline)

	// assert consistency
	stateHashFromState := optimisticStateAccess.StateCommitment()
	blockIndex := optimisticStateAccess.BlockIndex()
	if stateData.Commitment != stateHashFromState || blockIndex != task.AnchorOutput.StateIndex {
		// leaving earlier, state is not consistent and optimistic reader sync didn't catch it
		panic(coreutil.ErrorStateInvalidated)
	}
	openingStateUpdate := state.NewStateUpdateWithBlocklogValues(blockIndex+1, task.TimeAssumption.Time, stateData.Commitment)
	optimisticStateAccess.ApplyStateUpdates(openingStateUpdate)
	finalStateTimestamp := task.TimeAssumption.Time.Add(time.Duration(len(task.Requests)+1) * time.Nanosecond)

	ret := &VMContext{
		task:                 task,
		virtualState:         optimisticStateAccess,
		finalStateTimestamp:  finalStateTimestamp,
		blockContext:         make(map[iscp.Hname]*blockContext),
		blockContextCloseSeq: make([]iscp.Hname, 0),
		entropy:              task.Entropy,
		callStack:            make([]*callContext, 0),
	}
	nativeTokenBalanceLoader := func(id *iotago.NativeTokenID) (*iotago.ExtendedOutput, *iotago.UTXOInput) {
		return ret.loadNativeTokenOutput(id)
	}
	foundryLoader := func(serNum uint32) (*iotago.FoundryOutput, *iotago.UTXOInput) {
		return ret.loadFoundry(serNum)
	}
	ret.txbuilder = vmtxbuilder.NewAnchorTransactionBuilder(
		task.AnchorOutput,
		&task.AnchorOutputID,
		nativeTokenBalanceLoader,
		foundryLoader,
		task.RentStructure,
	)

	// at the beginning of each block we save the anchor tx ID of the current state
	if task.AnchorOutput.StateIndex > 0 {
		ret.currentStateUpdate = state.NewStateUpdate()
		ret.callCore(blocklog.Contract, func(s kv.KVStore) {
			blocklog.SetAnchorTransactionIDOfLatestBlock(s, ret.task.AnchorOutputID.TransactionID)
		})
		ret.virtualState.ApplyStateUpdates(ret.currentStateUpdate)
		ret.currentStateUpdate = nil
	}

	return ret
}

//nolint:revive
func (vmctx *VMContext) GetResult() (dict.Dict, error) {
	return vmctx.lastResult, vmctx.lastError
}

// CloseVMContext does the closing actions on the block
// return nil for normal block and rotation address for rotation block
func (vmctx *VMContext) CloseVMContext(numRequests, numSuccess, numOffLedger uint16) (uint32, hashing.HashValue, time.Time, iotago.Address) {
	vmctx.currentStateUpdate = state.NewStateUpdate() // need this before to make state valid
	rotationAddr := vmctx.saveBlockInfo(numRequests, numSuccess, numOffLedger)
	vmctx.closeBlockContexts()
	vmctx.saveInternalUTXOs()
	vmctx.virtualState.ApplyStateUpdates(vmctx.currentStateUpdate)

	blockIndex := vmctx.virtualState.BlockIndex()
	stateCommitment := vmctx.virtualState.StateCommitment()
	timestamp := vmctx.virtualState.Timestamp()

	return blockIndex, stateCommitment, timestamp, rotationAddr
}

func (vmctx *VMContext) checkRotationAddress() (ret iotago.Address) {
	vmctx.callCore(governance.Contract, func(s kv.KVStore) {
		ret = governance.GetRotationAddress(s)
	})
	return
}

// saveBlockInfo is in the blocklog partition context. Returns rotation address if this block is a rotation block
func (vmctx *VMContext) saveBlockInfo(numRequests, numSuccess, numOffLedger uint16) iotago.Address {
	if rotationAddress := vmctx.checkRotationAddress(); rotationAddress != nil {
		// block was marked fake by the governance contract because it is a committee rotation.
		// There was only on request in the block
		// We skip saving block information in order to avoid inconsistencies
		return rotationAddress
	}
	// block info will be stored into the separate state update
	prevStateData, err := iscp.StateDataFromBytes(vmctx.task.AnchorOutput.StateMetadata)
	if err != nil {
		panic(err)
	}
	dustAnchor, dustNativeToken := vmctx.txbuilder.DustDeposits()
	blockInfo := &blocklog.BlockInfo{
		BlockIndex:               vmctx.virtualState.BlockIndex(),
		Timestamp:                vmctx.virtualState.Timestamp(),
		TotalRequests:            numRequests,
		NumSuccessfulRequests:    numSuccess,
		NumOffLedgerRequests:     numOffLedger,
		PreviousStateHash:        prevStateData.Commitment,
		AnchorTransactionID:      iotago.TransactionID{}, // nil for now, will be updated the next round with the real tx id
		DustDepositAnchor:        dustAnchor,
		DustDepositNativeTokenID: dustNativeToken,
	}
	if vmctx.virtualState.PreviousStateHash() != blockInfo.PreviousStateHash {
		panic("CloseVMContext: inconsistent previous state hash")
	}

	vmctx.callCore(blocklog.Contract, func(s kv.KVStore) {
		idx := blocklog.SaveNextBlockInfo(s, blockInfo)
		if idx != blockInfo.BlockIndex {
			panic("CloseVMContext: inconsistent block index")
		}
		blocklog.SaveControlAddressesIfNecessary(
			s,
			vmctx.task.AnchorOutput.StateController,
			vmctx.task.AnchorOutput.GovernanceController,
			vmctx.task.AnchorOutput.StateIndex,
		)
	})
	return nil
}

// closeBlockContexts closing block contexts in deterministic FIFO sequence
func (vmctx *VMContext) closeBlockContexts() {
	for _, hname := range vmctx.blockContextCloseSeq {
		b := vmctx.blockContext[hname]
		b.onClose(b.obj)
	}
	vmctx.virtualState.ApplyStateUpdates(vmctx.currentStateUpdate)
}

func (vmctx *VMContext) saveInternalUTXOs() {
	nativeTokenIDs, nativeTokensToBeRemoved := vmctx.txbuilder.NativeTokenRecordsToBeUpdated()
	nativeTokensOutputsToBeUpdated := vmctx.txbuilder.NativeTokenOutputsByTokenIDs(nativeTokenIDs)

	foundryIDs, foundriesToBeRemoved := vmctx.txbuilder.FoundriesToBeUpdated()
	foundrySNToBeUpdated := vmctx.txbuilder.FoundryOutputsBySerNums(foundryIDs)

	blockIndex := vmctx.task.AnchorOutput.StateIndex + 1
	outputIndex := uint16(1)

	vmctx.callCore(accounts.Contract, func(s kv.KVStore) {
		// update native token outputs
		for _, out := range nativeTokensOutputsToBeUpdated {
			accounts.SaveNativeTokenOutput(s, out, blockIndex, outputIndex)
			outputIndex++
		}
		for _, id := range nativeTokensToBeRemoved {
			accounts.DeleteNativeTokenOutput(s, &id)
		}
		// update foundry UTXOs
		for _, out := range foundrySNToBeUpdated {
			accounts.SaveFoundryOutput(s, out, blockIndex, outputIndex)
			outputIndex++
		}
		for _, sn := range foundriesToBeRemoved {
			accounts.DeleteFoundryOutput(s, sn)
		}
	})
}
