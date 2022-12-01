package smGPA

import (
	"github.com/iotaledger/wasp/packages/chain/statemanager/smGPA/smInputs"
	"github.com/iotaledger/wasp/packages/state"
)

type stateBlockRequest struct {
	lastBlockHash  state.BlockHash
	lastBlockIndex uint32 // TODO: temporary field. Remove it after DB refactoring.
	priority       uint32
	done           bool
	blocks         []state.Block
	isValidFun     isValidFun
	respondFun     respondFun
}

type isValidFun func() bool

type respondFun func([]state.Block, state.VirtualStateAccess) // TODO: blocks parameter should probably be removed after DB refactoring

var _ blockRequest = &stateBlockRequest{}

func newStateBlockRequestTemplate() *stateBlockRequest {
	return &stateBlockRequest{
		done:   false,
		blocks: make([]state.Block, 0),
	}
}

func newStateBlockRequestFromConsensus(input *smInputs.ConsensusDecidedState) *stateBlockRequest {
	result := newStateBlockRequestTemplate()
	result.lastBlockHash = input.GetStateCommitment().BlockHash
	result.lastBlockIndex = input.GetBlockIndex()
	result.priority = topPriority
	result.isValidFun = func() bool {
		return input.IsValid()
	}
	result.respondFun = func(_ []state.Block, vState state.VirtualStateAccess) {
		input.Respond(vState)
	}
	return result
}

func newStateBlockRequestLocal(bi uint32, bh state.BlockHash, priority uint32, respondFun respondFun) *stateBlockRequest {
	result := newStateBlockRequestTemplate()
	result.lastBlockHash = bh
	result.lastBlockIndex = bi
	result.priority = priority
	result.isValidFun = func() bool {
		return true
	}
	result.respondFun = respondFun
	return result
}

func (sbrT *stateBlockRequest) getLastBlockHash() state.BlockHash {
	return sbrT.lastBlockHash
}

func (sbrT *stateBlockRequest) getLastBlockIndex() uint32 { // TODO: temporary function. Remove it after DB refactoring.
	return sbrT.lastBlockIndex
}

func (sbrT *stateBlockRequest) isValid() bool {
	if sbrT.done {
		return false
	}
	return sbrT.isValidFun()
}

func (sbrT *stateBlockRequest) getPriority() uint32 {
	return sbrT.priority
}

func (sbrT *stateBlockRequest) blockAvailable(block state.Block) {
	sbrT.blocks = append(sbrT.blocks, block)
}

func (sbrT *stateBlockRequest) markCompleted(createBaseStateFun createStateFun) {
	if sbrT.isValid() {
		sbrT.done = true
		baseState, err := createBaseStateFun()
		if err != nil {
			// Something failed in creating the base state. Just forget the request.
			return
		}
		if baseState == nil {
			// No need to respond. Just do nothing.
			return
		}
		vState := baseState
		for i := len(sbrT.blocks) - 1; i >= 0; i-- {
			calculatedStateCommitment := state.RootCommitment(vState.TrieNodeStore())
			if !state.EqualCommitments(calculatedStateCommitment, sbrT.blocks[i].PreviousL1Commitment().StateCommitment) {
				return
			}
			err := vState.ApplyBlock(sbrT.blocks[i])
			if err != nil {
				return
			}
			vState.Commit() // TODO: is it needed
		}
		sbrT.respondFun(sbrT.blocks, vState)
	}
}
