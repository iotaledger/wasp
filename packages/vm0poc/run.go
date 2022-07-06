package vm0poc

import (
	"fmt"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/trie.go/trie"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/util/panicutil"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/core/errors/coreerrors"
)

type VMRunner struct{}

type VMContext struct {
	task         *vm.VMTask
	virtualState state.VirtualStateAccess
}

const (
	ParamDeltaInt64 = "d"
	CounterStateVar = "counter"
)

func NewVMRunner() VMRunner {
	return VMRunner{}
}

func (r VMRunner) Run(task *vm.VMTask) {
	// optimistic read panic catcher for the whole VM task
	err := panicutil.CatchPanicReturnError(
		func() { runTask(task) },
		coreutil.ErrorStateInvalidated,
	)
	if err != nil {
		switch e := err.(type) {
		case *iscp.VMError:
			task.VMError = e
		case error:
			// May require a different error type here?
			task.VMError = coreerrors.ErrUntypedError.Create(e.Error())
		default:
			task.VMError = coreerrors.ErrUntypedError.Create(e.Error())
		}
		task.Log.Warnf("VM task has been abandoned due to invalidated state. ACS session id: %d", task.ACSSessionID)
	}
}

func runTask(task *vm.VMTask) {
	vmctx := createVMContext(task)
	oneRequestProcessed := false
	delta := int64(0)
	for _, req := range task.Requests {
		var skipReason string
		if oneRequestProcessed {
			skipReason = "only taking one request from the batch"
		} else {
			delta, skipReason = vmctx.getParam(req)
		}
		if skipReason != "" {
			task.Log.Infof("request skipped (ignored) by the VM: %s, reason: %s",
				req.ID().String(), skipReason)
			continue
		}

		result := vmctx.runTheRequest(delta)
		task.Log.Infof("request has been processed. %s, result: %s",
			req.ID().String(), result)
		oneRequestProcessed = true

		task.ResultTransactionEssence, task.ResultInputsCommitment = vmctx.closeVMContext()
		task.Results = []*vm.RequestResult{
			{
				Request: req,
				Return:  nil,
				Error:   nil,
				Receipt: nil,
			},
		}
	}
}

func (vmctx *VMContext) getParam(req iscp.Request) (int64, string) {
	if !req.IsOffLedger() {
		return 0, "only off-ledger requests are accepted"
	}
	p, err := req.Params().Get(ParamDeltaInt64)
	if err != nil {
		return 0, err.Error()
	}
	if p == nil {
		return 0, fmt.Sprintf("can't find param '%s'", ParamDeltaInt64)
	}
	delta, err := util.Int64From8Bytes(p)
	if err != nil {
		return 0, err.Error()
	}
	return delta, ""
}

func createVMContext(task *vm.VMTask) *VMContext {
	// assert consistency. It is a bit redundant double check
	if len(task.Requests) == 0 {
		// should never happen
		panic(fmt.Errorf("CreateVMContext.invalid params: must be at least 1 request"))
	}
	l1Commitment, err := state.L1CommitmentFromBytes(task.AnchorOutput.StateMetadata)
	if err != nil {
		// should never happen
		panic(fmt.Errorf("CreateVMContext: can't parse state data as L1Commitment from chain input %w", err))
	}
	// we create optimistic state access wrapper to be used inside the VM call.
	// It will panic any time the state is accessed.
	// The panic will be caught above and VM call will be abandoned peacefully
	optimisticStateAccess := state.WrapMustOptimisticVirtualStateAccess(task.VirtualStateAccess, task.SolidStateBaseline)

	// assert consistency
	commitmentFromState := state.RootCommitment(optimisticStateAccess.TrieNodeStore())
	blockIndex := optimisticStateAccess.BlockIndex()
	if !state.EqualCommitments(l1Commitment.StateCommitment, commitmentFromState) || blockIndex != task.AnchorOutput.StateIndex {
		// leaving earlier, state is not consistent and optimistic reader sync didn't catch it
		panic(coreutil.ErrorStateInvalidated)
	}
	openingStateUpdate := state.NewStateUpdateWithBlockLogValues(blockIndex+1, task.TimeAssumption.Time, &l1Commitment)
	optimisticStateAccess.ApplyStateUpdate(openingStateUpdate)

	ret := &VMContext{
		task:         task,
		virtualState: optimisticStateAccess,
	}
	return ret
}

func (vmctx *VMContext) runTheRequest(delta int64) (result string) {
	var c int64

	if vmctx.task.AnchorOutput.StateIndex == 0 {
		vmctx.virtualState.KVStore().Set(CounterStateVar, util.Int64To8Bytes(delta))
	} else {
		b, err := vmctx.virtualState.KVStore().Get(CounterStateVar)
		if err != nil {
			panic(err)
		}
		c, err = util.Int64From8Bytes(b)
		if err != nil {
			panic(err)
		}
		vmctx.virtualState.KVStore().Set(CounterStateVar, util.Int64To8Bytes(c+delta))
	}
	if !vmctx.task.SolidStateBaseline.IsValid() {
		panic(coreutil.ErrorStateInvalidated)
	}
	return fmt.Sprintf("counter was mutated: %d + (%d) -> %d", c, delta, c+delta)
}

func (vmctx *VMContext) closeVMContext() (*iotago.TransactionEssence, []byte) {
	block, err := vmctx.virtualState.ExtractBlock()
	if err != nil {
		panic(err)
	}
	vmctx.virtualState.Commit()
	stateCommitment := trie.RootCommitment(vmctx.virtualState.TrieNodeStore())
	blockHash := state.BlockHashFromData(block.EssenceBytes())
	l1Commitment := state.NewL1Commitment(stateCommitment, blockHash)

	anchor := vmctx.task.AnchorOutput.Clone().(*iotago.AliasOutput)
	anchor.StateMetadata = l1Commitment.Bytes()
	aliasID := vmctx.task.AnchorOutput.AliasID
	if aliasID.Empty() {
		aliasID = iotago.AliasIDFromOutputID(vmctx.task.AnchorOutputID)
	}
	anchor.AliasID = aliasID
	anchor.Features = iotago.Features{
		&iotago.SenderFeature{
			Address: aliasID.ToAddress(),
		},
	}
	anchor.StateIndex = vmctx.task.AnchorOutput.StateIndex + 1

	inputIDs := iotago.OutputIDs{vmctx.task.AnchorOutputID}
	inputSet := iotago.OutputSet{vmctx.task.AnchorOutputID: vmctx.task.AnchorOutput}
	inputs := inputIDs.UTXOInputs()
	essence := &iotago.TransactionEssence{
		NetworkID: parameters.L1.Protocol.NetworkID(),
		Inputs:    inputs,
		Outputs:   iotago.Outputs{anchor},
	}

	inputsCommitment := inputIDs.OrderedSet(inputSet).MustCommitment()
	copy(essence.InputsCommitment[:], inputsCommitment)

	return essence, inputsCommitment
}
