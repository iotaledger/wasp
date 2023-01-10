package smGPA

import (
	"fmt"

	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/chain/statemanager/smGPA/smInputs"
	"github.com/iotaledger/wasp/packages/state"
)

type blockRequestImpl struct {
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

var _ blockRequest = &blockRequestImpl{}

func newBlockRequestTemplate(rType string, log *logger.Logger, id blockRequestID) *blockRequestImpl {
	return &blockRequestImpl{
		done:   false,
		blocks: make([]state.Block, 0),
		log:    log.Named(fmt.Sprintf("r%s-%v", rType, id)),
		rType:  rType,
		id:     id,
	}
}

func newBlockRequestFromConsensusStateProposal(input *smInputs.ConsensusStateProposal, log *logger.Logger, id blockRequestID) *blockRequestImpl {
	result := newBlockRequestTemplate("scsp", log, id)
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

func newBlockRequestFromConsensusDecidedState(input *smInputs.ConsensusDecidedState, log *logger.Logger, id blockRequestID) *blockRequestImpl {
	result := newBlockRequestTemplate("scds", log, id)
	result.lastL1Commitment = input.GetL1Commitment()
	result.isValidFun = func() bool {
		return input.IsValid()
	}
	result.respondFun = func(obtainStateFun obtainStateFun) {
		state, err := obtainStateFun()
		if err != nil {
			result.log.Errorf("error obtaining state: %w", err)
			return
		}
		input.Respond(state)
	}
	result.log.Debugf("State block request from consensus decided state id %v for block %s is created", result.id, result.lastL1Commitment)
	return result
}

func newBlockRequestFromMempool(typeStr string, commitment *state.L1Commitment, isValidFun isValidFun, respondFun respondFun, log *logger.Logger, id blockRequestID) *blockRequestImpl {
	result := newBlockRequestTemplate("sm"+typeStr, log, id)
	result.lastL1Commitment = commitment
	result.isValidFun = isValidFun
	result.respondFun = respondFun
	result.log.Debugf("State block request from mempool type %s id %v for block %s is created",
		typeStr, result.id, result.lastL1Commitment)
	return result
}

func newBlockRequestLocal(commitment *state.L1Commitment, respondFun respondFun, log *logger.Logger, id blockRequestID) *blockRequestImpl {
	result := newBlockRequestTemplate("sl", log, id)
	result.lastL1Commitment = commitment
	result.isValidFun = func() bool {
		return true
	}
	result.respondFun = respondFun
	result.log.Debugf("Local state block request id %v for block %s is created", result.id, result.lastL1Commitment)
	return result
}

func (sbrT *blockRequestImpl) getLastL1Commitment() *state.L1Commitment {
	return sbrT.lastL1Commitment
}

func (sbrT *blockRequestImpl) isValid() bool {
	if sbrT.done {
		return false
	}
	return sbrT.isValidFun()
}

func (sbrT *blockRequestImpl) blockAvailable(block state.Block) {
	sbrT.log.Debugf("State block request received block %s, appending it to chain", block.L1Commitment())
	sbrT.blocks = append(sbrT.blocks, block)
}

func (sbrT *blockRequestImpl) getBlockChain() []state.Block {
	return sbrT.blocks
}

func (sbrT *blockRequestImpl) getChainOfBlocks(baseIndex uint32, obtainBlockFun obtainBlockFun) chainOfBlocks {
	return newChainOfBlocks(sbrT.getBlockChain(), sbrT.getLastL1Commitment(), baseIndex, obtainBlockFun, sbrT.log)
}

func (sbrT *blockRequestImpl) markCompleted(obtainStateFun obtainStateFun) {
	if sbrT.isValid() {
		sbrT.log.Debugf("State block request is valid, marking it completed and responding")
		sbrT.done = true
		sbrT.respondFun(obtainStateFun)
		sbrT.log.Debugf("Responded")
	} else {
		sbrT.log.Debugf("State block request is not valid, ignoring mark completed call")
	}
}

func (sbrT *blockRequestImpl) getType() string {
	return sbrT.rType
}

func (sbrT *blockRequestImpl) getID() blockRequestID {
	return sbrT.id
}
