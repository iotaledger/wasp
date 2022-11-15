package smGPA

import (
	"fmt"

	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/chain/statemanager/smGPA/smInputs"
	"github.com/iotaledger/wasp/packages/state"
)

type stateBlockRequest struct {
	lastBlockHash  state.BlockHash
	lastBlockIndex uint32 // TODO: temporary field. Remove it after DB refactoring.
	priority       uint64
	done           bool
	blocks         []state.Block
	isValidFun     isValidFun
	respondFun     respondFun
	log            *logger.Logger
	rType          string
	id             blockRequestID
}

type isValidFun func() bool

type respondFun func([]state.Block, state.VirtualStateAccess) // TODO: blocks parameter should probably be removed after DB refactoring

var _ blockRequest = &stateBlockRequest{}

func newStateBlockRequestTemplate(rType string, log *logger.Logger, id blockRequestID) *stateBlockRequest {
	return &stateBlockRequest{
		done:   false,
		blocks: make([]state.Block, 0),
		log:    log.Named(fmt.Sprintf("r%s-%v", rType, id)),
		rType:  rType,
		id:     id,
	}
}

func newStateBlockRequestFromConsensus(input *smInputs.ConsensusDecidedState, log *logger.Logger, id blockRequestID) *stateBlockRequest {
	result := newStateBlockRequestTemplate("scds", log, id)
	result.lastBlockHash = input.GetStateCommitment().BlockHash
	result.lastBlockIndex = input.GetBlockIndex()
	result.priority = topPriority
	result.isValidFun = func() bool {
		return input.IsValid()
	}
	result.respondFun = func(_ []state.Block, vState state.VirtualStateAccess) {
		input.Respond(vState)
	}
	result.log.Debugf("State block request from consensus id %v for block %s is created", result.id, result.lastBlockHash)
	return result
}

func newStateBlockRequestLocal(bi uint32, bh state.BlockHash, respondFun respondFun, log *logger.Logger, id blockRequestID) *stateBlockRequest {
	result := newStateBlockRequestTemplate("sl", log, id)
	result.lastBlockHash = bh
	result.lastBlockIndex = bi
	result.priority = uint64(id)
	result.isValidFun = func() bool {
		return true
	}
	result.respondFun = respondFun
	result.log.Debugf("Local state block request id %v for block %s is created", result.id, result.lastBlockHash)
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

func (sbrT *stateBlockRequest) getPriority() uint64 {
	return sbrT.priority
}

func (sbrT *stateBlockRequest) blockAvailable(block state.Block) {
	sbrT.log.Debugf("State block request received block %s, appending it to chain", block.GetHash())
	sbrT.blocks = append(sbrT.blocks, block)
}

func (sbrT *stateBlockRequest) markCompleted(createBaseStateFun createStateFun) {
	if sbrT.isValid() {
		sbrT.log.Debugf("State block request is valid, marking it completed: calculating state")
		sbrT.done = true
		baseState, err := createBaseStateFun()
		if err != nil {
			sbrT.log.Errorf("Error creating base state: %v", err)
			return
		}
		if baseState == nil {
			sbrT.log.Debugf("Created base state is nil: skipping responding to the request")
			return
		}
		vState := baseState
		for i := len(sbrT.blocks) - 1; i >= 0; i-- {
			calculatedStateCommitment := state.RootCommitment(vState.TrieNodeStore())
			if !state.EqualCommitments(calculatedStateCommitment, sbrT.blocks[i].PreviousL1Commitment().StateCommitment) {
				sbrT.log.Errorf("State index %v root commitment does not match block %s expected commitment", vState.BlockIndex(), sbrT.blocks[i].GetHash())
				return
			}
			err := vState.ApplyBlock(sbrT.blocks[i])
			if err != nil {
				sbrT.log.Errorf("Error applying block %s to state index %v: %v", sbrT.blocks[i].GetHash(), vState.BlockIndex(), err)
				return
			}
			vState.Commit() // TODO: is it needed
		}
		sbrT.log.Debugf("State index %v is calculated, responding", vState.BlockIndex())
		sbrT.respondFun(sbrT.blocks, vState)
		sbrT.log.Debugf("Responded")
	} else {
		sbrT.log.Debugf("State block request is not valid, ignoring mark completed call")
	}
}

func (sbrT *stateBlockRequest) getType() string {
	return sbrT.rType
}

func (sbrT *stateBlockRequest) getID() blockRequestID {
	return sbrT.id
}
