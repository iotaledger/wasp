package vmcontext

import (
	"time"

	"github.com/iotaledger/wasp/packages/transaction"

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

var (
	NewSandbox     func(vmctx *VMContext) iscp.Sandbox
	NewSandboxView func(vmctx *VMContext) iscp.SandboxView
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
	gasBurnedTotal       uint64
	gasFeeChargedTotal   uint64

	// ---- request context
	chainInfo          *governance.ChainInfo
	req                iscp.Request
	requestIndex       uint16
	requestEventIndex  uint16
	currentStateUpdate state.StateUpdate
	entropy            hashing.HashValue
	contractRecord     *root.ContractRecord
	callStack          []*callContext
	// --- gas related
	// max tokens available for gas fee
	gasMaxTokensAvailableForGasFee uint64
	// final gas budget set for the run
	gasBudget uint64
	// is gas bur enabled
	gasBurnEnabled bool
	// gas already burned
	gasBurned uint64
	// tokens charged
	gasFeeCharged uint64
}

type callContext struct {
	caller             *iscp.AgentID // calling agent
	contract           iscp.Hname    // called contract
	params             dict.Dict     // params passed
	allowanceAvailable *iscp.Assets  // MUTABLE: allowance budget left after TransferAllowedFunds
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

	// at the beginning of each block
	var dustAssumptions *transaction.DustDepositAssumption

	if task.AnchorOutput.StateIndex > 0 {
		ret.currentStateUpdate = state.NewStateUpdate()

		// load and validate chain's dust assumptions about internal outputs. They must not get bigger!
		ret.callCore(accounts.Contract, func(s kv.KVStore) {
			dustAssumptions = accounts.GetDustAssumptions(s)
		})
		currentDustDepositValues := transaction.NewDepositEstimate(task.RentStructure)
		if currentDustDepositValues.AnchorOutput > dustAssumptions.AnchorOutput ||
			currentDustDepositValues.NativeTokenOutput > dustAssumptions.NativeTokenOutput {
			panic(ErrInconsistentDustAssumptions)
		}

		// save the anchor tx ID of the current state
		ret.callCore(blocklog.Contract, func(s kv.KVStore) {
			blocklog.SetAnchorTransactionIDOfLatestBlock(s, ret.task.AnchorOutputID.TransactionID)
		})

		ret.virtualState.ApplyStateUpdates(ret.currentStateUpdate)
		ret.currentStateUpdate = nil
	} else {
		// assuming dust assumptions for the first block. It must be consistent with parameters in the init request
		dustAssumptions = transaction.NewDepositEstimate(task.RentStructure)
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
		*dustAssumptions,
		task.RentStructure,
	)

	return ret
}

// CloseVMContext does the closing actions on the block
// return nil for normal block and rotation address for rotation block
func (vmctx *VMContext) CloseVMContext(numRequests, numSuccess, numOffLedger uint16) (uint32, hashing.HashValue, time.Time, iotago.Address) {
	vmctx.gasBurnEnable(false)
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
	totalIotasInContracts, totalDustOnChain := vmctx.txbuilder.TotalIotasInOutputs()
	blockInfo := &blocklog.BlockInfo{
		BlockIndex:             vmctx.virtualState.BlockIndex(),
		Timestamp:              vmctx.virtualState.Timestamp(),
		TotalRequests:          numRequests,
		NumSuccessfulRequests:  numSuccess,
		NumOffLedgerRequests:   numOffLedger,
		PreviousStateHash:      prevStateData.Commitment,
		AnchorTransactionID:    iotago.TransactionID{}, // nil for now, will be updated the next round with the real tx id
		TotalIotasInL2Accounts: totalIotasInContracts,
		TotalDustDeposit:       totalDustOnChain,
		GasBurned:              vmctx.gasBurnedTotal,
		GasFeeCharged:          vmctx.gasFeeChargedTotal,
	}
	if vmctx.virtualState.PreviousStateHash() != blockInfo.PreviousStateHash {
		panic("CloseVMContext: inconsistent previous state hash")
	}

	vmctx.callCore(blocklog.Contract, func(s kv.KVStore) {
		blocklog.SaveNextBlockInfo(s, blockInfo)
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
	foundrySNToBeUpdated := vmctx.txbuilder.FoundryOutputsBySN(foundryIDs)

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

func (vmctx *VMContext) assertConsistentL2WithL1TxBuilder(checkpoint string) {
	if vmctx.task.AnchorOutput.StateIndex == 0 && vmctx.isInitChainRequest() {
		return
	}
	var totalL2Assets *iscp.Assets
	vmctx.callCore(accounts.Contract, func(s kv.KVStore) {
		totalL2Assets = accounts.GetTotalL2Assets(s)
	})
	vmctx.txbuilder.AssertConsistentWithL2Totals(totalL2Assets, checkpoint)
}
