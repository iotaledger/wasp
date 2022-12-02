package smGPA

import (
	//	"fmt"

	//"github.com/iotaledger/hive.go/core/logger"
	//	"github.com/iotaledger/wasp/packages/chain/statemanager/smGPA/smInputs"
	"github.com/iotaledger/wasp/packages/state"
)

/*type requestCommonAncestorImpl struct {
	input                         *smInputs.MempoolStateRequest
	oldStateBlockRequest          *stateBlockRequest
	newStateBlockRequest          *stateBlockRequest
	oldStateBlockRequestCompleted bool
	newStateBlockRequestCompleted bool
	obtainNewStateFun             obtainStateFun
	obtainCommittedBlockFun       obtainBlockFun
	log                           *logger.Logger
}*/

type obtainBlockFun func(*state.L1Commitment) state.Block

/*var _ requestCommonAncestor = &requestCommonAncestorImpl{}

func newCommonAncestorBlockRequest(input *smInputs.MempoolStateRequest, ocbFun obtainBlockFun, log *logger.Logger, id blockRequestID) requestCommonAncestor {
	result := &requestCommonAncestorImpl{
		input:                         input,
		oldStateBlockRequestCompleted: false,
		newStateBlockRequestCompleted: false,
		obtainCommittedBlockFun:       ocbFun,
		log:                           log.Named(fmt.Sprintf("rca-%v", id)),
	}
	result.oldStateBlockRequest = newStateBlockRequestFromMempool(result, mempoolStateBlockRequestTypeOld, result.log, id)
	result.newStateBlockRequest = newStateBlockRequestFromMempool(result, mempoolStateBlockRequestTypeNew, result.log, id)
	return result
}

func (rcaiT *requestCommonAncestorImpl) getInput() *smInputs.MempoolStateRequest {
	return rcaiT.input
}

func (rcaiT *requestCommonAncestorImpl) isValid() bool {
	return rcaiT.getInput().IsValid()
}*/

/*func (rcaiT *requestCommonAncestorImpl) blockAvailable(block state.Block, blockIndex uint32, msbrType byte) {
	var other *mempoolStateBlockRequest
	switch msbrType {
	case mempoolStateBlockRequestTypeOld:
		other = rcaiT.newStateBlockRequest
	case mempoolStateBlockRequestTypeNew:
		other = rcaiT.oldStateBlockRequest
	default:
		rcaiT.log.Panicf("Unknown mempool state request type %v", msbrType)
	}

}*/

/*func (rcaiT *requestCommonAncestorImpl) stateRequestCompleted(obtainStateFun obtainStateFun, msbrType byte) {
	switch msbrType {
	case mempoolStateBlockRequestTypeOld:
		rcaiT.oldStateBlockRequestCompleted = true
	case mempoolStateBlockRequestTypeNew:
		rcaiT.newStateBlockRequestCompleted = true
		rcaiT.obtainNewStateFun = obtainStateFun
	default:
		rcaiT.log.Panicf("Unknown mempool state request type %v", msbrType)
	}

	if !rcaiT.oldStateBlockRequestCompleted || !rcaiT.newStateBlockRequestCompleted {
		return
	}

	oldBaseIndex := rcaiT.getInput().GetOldStateIndex()
	newBaseIndex := rcaiT.getInput().GetNewStateIndex()
	var commonIndex uint32
	if newBaseIndex > oldBaseIndex {
		commonIndex = oldBaseIndex
	} else {
		commonIndex = newBaseIndex
	}

	oldBC := rcaiT.oldStateBlockRequest.getBlockChain()
	newBC := rcaiT.newStateBlockRequest.getBlockChain()
	oldCommitment := rcaiT.getInput().GetOldL1Commitment()
	newCommitment := rcaiT.getInput().GetNewL1Commitment()
	oldCOB := newChainOfBlocks(oldBC, oldCommitment, oldBaseIndex, rcaiT.obtainCommittedBlockFun, rcaiT.log.Named("old"))
	newCOB := newChainOfBlocks(newBC, newCommitment, newBaseIndex, rcaiT.obtainCommittedBlockFun, rcaiT.log.Named("new"))

	respondFun := func(index uint32) {
		newState, err := rcaiT.obtainNewStateFun()
		if err != nil {
			rcaiT.log.Errorf("Unable to obtain new state: %v", err)
			return
		}

		rcaiT.input.Respond(smInputs.NewMempoolStateRequestResults(
			newState,
			newCOB.getBlocksFrom(index),
			oldCOB.getBlocksFrom(index),
		))
	}

	for commonIndex > 0 {
		if oldCOB.getL1Commitment(commonIndex).Equals(newCOB.getL1Commitment(commonIndex)) {
			respondFun(commonIndex)
			return
		}
		commonIndex--
	}
	respondFun(0)
}*/

// chainOfBlocks is used in requestCommonAncestorImpl only

type chainOfBlocks struct {
	blocks                  []state.Block
	baseCommitment          *state.L1Commitment
	baseIndex               uint32
	obtainCommittedBlockFun obtainBlockFun
}

func newChainOfBlocks(blocks []state.Block, baseCommitment *state.L1Commitment, baseIndex uint32, ocbFun obtainBlockFun) *chainOfBlocks {
	return &chainOfBlocks{
		blocks:                  blocks,
		baseCommitment:          baseCommitment,
		baseIndex:               baseIndex,
		obtainCommittedBlockFun: ocbFun,
	}
}

func (cobT *chainOfBlocks) getL1Commitment(blockIndex uint32) *state.L1Commitment {
	index := cobT.baseIndex - blockIndex
	if index < uint32(len(cobT.blocks)) {
		return cobT.blocks[index].L1Commitment()
	}
	var previousCommitment *state.L1Commitment
	if len(cobT.blocks) == 0 {
		previousCommitment = cobT.baseCommitment
	} else {
		previousCommitment = cobT.blocks[len(cobT.blocks)-1].PreviousL1Commitment()
	}
	for i := len(cobT.blocks); uint32(i) <= index; i++ {
		block := cobT.obtainCommittedBlockFun(previousCommitment)
		cobT.blocks = append(cobT.blocks, block)
		previousCommitment = block.PreviousL1Commitment()
	}
	return cobT.blocks[index].L1Commitment()
}

func (cobT *chainOfBlocks) getBlocksFrom(blockIndex uint32) []state.Block { // Not including index; in proper order
	result := make([]state.Block, cobT.baseIndex-blockIndex)
	for i := range result {
		result[i] = cobT.blocks[cobT.baseIndex-blockIndex-uint32(i)-1]
	}
	return result
}
