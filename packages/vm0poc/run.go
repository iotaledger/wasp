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

// VMRunner minimal mocked VM. It receives requests. Each request is expected to have
// parameter ParamDeltaInt64 = "d" with the delta value
// The request updates the state of the chain by adding `delta` to the int64 value stored at
// the key "counter"
// The VM takes any batch of requests but only processes one single off-ledger request which it is able
// to parse successfully. Other requests are skipped, which means they will be picked in the next block
// Warning: probably not 100% correct in edge cases
type VMRunner struct{}

type VMContext struct {
	task         *vm.VMTask
	virtualState state.VirtualStateAccess
}

const (
	// ParamDeltaInt64 name of the parameter in the request
	ParamDeltaInt64 = "d"
	// CounterStateVar key of counter in the state
	CounterStateVar = "counter"
)

func NewVMRunner() VMRunner {
	return VMRunner{}
}

func (r VMRunner) Run(task *vm.VMTask) {
	// optimistic read panic catcher for the whole VM task
	// If the node updates the state while VM is still running, it will peacefully
	// abandon the VM run
	err := panicutil.CatchPanicReturnError(
		func() {
			runTask(task)
		},
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

// runTask initialized, loops through batch and finalizes by producing updated (not committed state)
// and the anchor transaction
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
		task.Results = []*vm.RequestResult{{Request: req}}
	}
}

// getParam parses the request
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
	// It will panic any time the state is write-accessed.
	// The panic will be caught above and VM call will be abandoned peacefully
	optimisticStateAccess := state.WrapMustOptimisticVirtualStateAccess(task.VirtualStateAccess, task.SolidStateBaseline)

	// assert consistency
	commitmentFromState := state.RootCommitment(optimisticStateAccess.TrieNodeStore())
	blockIndex := optimisticStateAccess.BlockIndex()
	if !state.EqualCommitments(l1Commitment.StateCommitment, commitmentFromState) || blockIndex != task.AnchorOutput.StateIndex {
		// leaving earlier, state is not consistent and optimistic reader sync didn't catch it
		panic(coreutil.ErrorStateInvalidated)
	}
	openingStateUpdate := state.NewStateUpdateWithBlockLogValues(
		blockIndex+1,
		task.TimeAssumption,
		&l1Commitment,
	)
	optimisticStateAccess.ApplyStateUpdate(openingStateUpdate)

	ret := &VMContext{
		task:         task,
		virtualState: optimisticStateAccess,
	}
	return ret
}

// runTheRequest updates the state
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

// closeVMContext:
// - extracts mutation buffer as block
// - commits the trie (does not save the state)
// - produces next anchor output
// - produces anchor transaction essence
func (vmctx *VMContext) closeVMContext() (*iotago.TransactionEssence, []byte) {
	// block is just collection of mutations to the state in the form of Set(k,v) and Del(k)
	block, err := vmctx.virtualState.ExtractBlock()
	if err != nil {
		panic(err)
	}
	// recalculate the trie, in order to have the root commitment corresponding to the updated state
	// Note: this step does not save anything, only loads necessary nodes of the trie from DB to
	// the buffer and updates commitments in the nodes
	vmctx.virtualState.Commit()
	// take the root commitment of the updated state (not saved yet)
	stateCommitment := trie.RootCommitment(vmctx.virtualState.TrieNodeStore())
	// take deterministic hash of the mutations (plain hash, no merklization)
	blockHash := state.BlockHashFromData(block.EssenceBytes())
	// compose L1Commitment structure which will be put into the anchor Alias output
	l1Commitment := state.NewL1Commitment(stateCommitment, blockHash)

	// create new anchor Alias output. First cloning the previous one and then updating it
	nextAnchorOutput := vmctx.task.AnchorOutput.Clone().(*iotago.AliasOutput)
	// place l1commitment into the StateMetadata block
	nextAnchorOutput.StateMetadata = l1Commitment.Bytes()
	if vmctx.task.AnchorOutput.AliasID.Empty() {
		// if aliasID is all-0, it means the previous state was origin and the aliasID is not yet calculated.
		// We need to compute real AliasID from the outputID and put it into the output

		aliasID := iotago.AliasIDFromOutputID(vmctx.task.AnchorOutputID)
		nextAnchorOutput.Features = iotago.Features{
			&iotago.SenderFeature{
				Address: aliasID.ToAddress(),
			},
		}
		nextAnchorOutput.AliasID = aliasID
	}
	// increment state index
	nextAnchorOutput.StateIndex = vmctx.task.AnchorOutput.StateIndex + 1

	// compose the essence
	inputIDs := iotago.OutputIDs{vmctx.task.AnchorOutputID}
	inputSet := iotago.OutputSet{vmctx.task.AnchorOutputID: vmctx.task.AnchorOutput}
	inputs := inputIDs.UTXOInputs()
	essence := &iotago.TransactionEssence{
		NetworkID: parameters.L1.Protocol.NetworkID(),
		Inputs:    inputs,
		Outputs:   iotago.Outputs{nextAnchorOutput},
	}

	// set the commitment to real inputs: required by the Stardust to prevent fake output attacks
	inputsCommitment := inputIDs.OrderedSet(inputSet).MustCommitment()
	copy(essence.InputsCommitment[:], inputsCommitment)

	return essence, inputsCommitment
}
