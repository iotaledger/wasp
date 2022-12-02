package smGPA

import (
	"fmt"

	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/chain/statemanager/smGPA/smInputs"
	"github.com/iotaledger/wasp/packages/state"
)

type stateBlockRequest struct {
	lastL1Commitment *state.L1Commitment
	done             bool
	blocks           []state.Block
	isValidFun       isValidFun
	respondFun       respondFun
	log              *logger.Logger
	rType            string
	id               blockRequestID
}

type isValidFun func() bool

type respondFun func(obtainStateFun)

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

func newStateBlockRequestFromConsensusStateProposal(input *smInputs.ConsensusStateProposal, log *logger.Logger, id blockRequestID) *stateBlockRequest {
	result := newStateBlockRequestTemplate("scsp", log, id)
	result.lastL1Commitment = input.GetL1Commitment()
	result.isValidFun = func() bool {
		return input.IsValid()
	}
	result.respondFun = func(obtainStateFun obtainStateFun) {
		input.Respond()
	}
	result.log.Debugf("State block request from consensus state proposal id %v for block %s is created", result.id, result.lastL1Commitment)
	return result
}

func newStateBlockRequestFromConsensusDecidedState(input *smInputs.ConsensusDecidedState, log *logger.Logger, id blockRequestID) *stateBlockRequest {
	result := newStateBlockRequestTemplate("scds", log, id)
	result.lastL1Commitment = input.GetL1Commitment()
	result.isValidFun = func() bool {
		return input.IsValid()
	}
	result.respondFun = func(obtainStateFun obtainStateFun) {
		state, err := obtainStateFun()
		if err != nil {
			result.log.Errorf("Error obtaining state: %v", err)
			return
		}
		input.Respond(state)
	}
	result.log.Debugf("State block request from consensus decided state id %v for block %s is created", result.id, result.lastL1Commitment)
	return result
}

func newStateBlockRequestFromMempool(typeStr string, commitment *state.L1Commitment, isValidFun isValidFun, respondFun respondFun, log *logger.Logger, id blockRequestID) *stateBlockRequest {
	result := newStateBlockRequestTemplate("sm"+typeStr, log, id)
	result.lastL1Commitment = commitment
	result.isValidFun = isValidFun
	result.respondFun = respondFun
	result.log.Debugf("State block request from mempool type %s id %v for block %s is created",
		typeStr, result.id, result.lastL1Commitment)
	return result
}

func newStateBlockRequestLocal(commitment *state.L1Commitment, respondFun respondFun, log *logger.Logger, id blockRequestID) *stateBlockRequest {
	result := newStateBlockRequestTemplate("sl", log, id)
	result.lastL1Commitment = commitment
	result.isValidFun = func() bool {
		return true
	}
	result.respondFun = respondFun
	result.log.Debugf("Local state block request id %v for block %s is created", result.id, result.lastL1Commitment)
	return result
}

func (sbrT *stateBlockRequest) getLastL1Commitment() *state.L1Commitment {
	return sbrT.lastL1Commitment
}

func (sbrT *stateBlockRequest) isValid() bool {
	if sbrT.done {
		return false
	}
	return sbrT.isValidFun()
}

func (sbrT *stateBlockRequest) blockAvailable(block state.Block) {
	sbrT.log.Debugf("State block request received block %s, appending it to chain", block.L1Commitment())
	sbrT.blocks = append(sbrT.blocks, block)
}

func (sbrT *stateBlockRequest) getBlockChain() []state.Block {
	return sbrT.blocks
}

func (sbrT *stateBlockRequest) markCompleted(obtainStateFun obtainStateFun) {
	if sbrT.isValid() {
		sbrT.log.Debugf("State block request is valid, marking it completed and responding")
		sbrT.done = true
		sbrT.respondFun(obtainStateFun)
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
