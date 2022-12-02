package smGPA

import (
	"github.com/iotaledger/wasp/packages/state"
)

type chainOfBlocksImpl struct {
	blocks                  []state.Block
	baseCommitment          *state.L1Commitment
	baseIndex               uint32
	obtainCommittedBlockFun obtainBlockFun
}

var _ chainOfBlocks = &chainOfBlocksImpl{}

func newChainOfBlocks(blocks []state.Block, baseCommitment *state.L1Commitment, baseIndex uint32, ocbFun obtainBlockFun) chainOfBlocks {
	return &chainOfBlocksImpl{
		blocks:                  blocks,
		baseCommitment:          baseCommitment,
		baseIndex:               baseIndex,
		obtainCommittedBlockFun: ocbFun,
	}
}

func (cobiT *chainOfBlocksImpl) getL1Commitment(blockIndex uint32) *state.L1Commitment {
	index := cobiT.baseIndex - blockIndex
	if index < uint32(len(cobiT.blocks)) {
		return cobiT.blocks[index].L1Commitment()
	}
	var previousCommitment *state.L1Commitment
	if len(cobiT.blocks) == 0 {
		previousCommitment = cobiT.baseCommitment
	} else {
		previousCommitment = cobiT.blocks[len(cobiT.blocks)-1].PreviousL1Commitment()
	}
	for i := len(cobiT.blocks); uint32(i) <= index; i++ {
		block := cobiT.obtainCommittedBlockFun(previousCommitment)
		cobiT.blocks = append(cobiT.blocks, block)
		previousCommitment = block.PreviousL1Commitment()
	}
	return cobiT.blocks[index].L1Commitment()
}

func (cobiT *chainOfBlocksImpl) getBlocksFrom(blockIndex uint32) []state.Block {
	result := make([]state.Block, cobiT.baseIndex-blockIndex)
	for i := range result {
		result[i] = cobiT.blocks[cobiT.baseIndex-blockIndex-uint32(i)-1]
	}
	return result
}
