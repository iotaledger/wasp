package smGPA

import (
	"github.com/iotaledger/hive.go/core/logger"
	"github.com/iotaledger/wasp/packages/state"
)

type chainOfBlocksImpl struct {
	blocks                  []state.Block
	baseCommitment          *state.L1Commitment
	baseIndex               uint32
	obtainCommittedBlockFun obtainBlockFun
	log                     *logger.Logger
}

var _ chainOfBlocks = &chainOfBlocksImpl{}

func newChainOfBlocks(blocks []state.Block, baseCommitment *state.L1Commitment, baseIndex uint32, ocbFun obtainBlockFun, log *logger.Logger) chainOfBlocks {
	return &chainOfBlocksImpl{
		blocks:                  blocks,
		baseCommitment:          baseCommitment,
		baseIndex:               baseIndex,
		obtainCommittedBlockFun: ocbFun,
		log:                     log.Named("cob"),
	}
}

func (cobiT *chainOfBlocksImpl) getL1Commitment(blockIndex uint32) *state.L1Commitment {
	cobiT.log.Debugf("Getting block %v commitment: total blocks: %v, base index: %v...", blockIndex, len(cobiT.blocks), cobiT.baseIndex)
	index := cobiT.baseIndex - blockIndex
	if index < uint32(len(cobiT.blocks)) {
		result := cobiT.blocks[index].L1Commitment()
		cobiT.log.Debugf("Getting block %v commitment: block is already obtained, commitment is %v", blockIndex, result)
		return result
	}
	var previousCommitment *state.L1Commitment
	if len(cobiT.blocks) == 0 {
		previousCommitment = cobiT.baseCommitment
		cobiT.log.Debugf("Getting block %v commitment: block is not yet obtained, starting from base block %s", blockIndex, previousCommitment)
	} else {
		previousCommitment = cobiT.blocks[len(cobiT.blocks)-1].PreviousL1Commitment()
		cobiT.log.Debugf("Getting block %v commitment: block is not yet obtained, starting from last block's previous block %s", blockIndex, previousCommitment)
	}
	for i := len(cobiT.blocks); uint32(i) <= index; i++ {
		cobiT.log.Debugf("Getting block %v commitment: obtaining block %s", blockIndex, previousCommitment)
		block := cobiT.obtainCommittedBlockFun(previousCommitment)
		cobiT.blocks = append(cobiT.blocks, block)
		previousCommitment = block.PreviousL1Commitment()
	}
	result := cobiT.blocks[index].L1Commitment()
	cobiT.log.Debugf("Getting block %v commitment: block obtained, commitment: %s", blockIndex, result)
	return result
}

func (cobiT *chainOfBlocksImpl) getBlocksFrom(blockIndex uint32) []state.Block {
	cobiT.log.Debugf("Getting blocks from index %v: total blocks: %v, base index: %v...", blockIndex, len(cobiT.blocks), cobiT.baseIndex)
	result := make([]state.Block, cobiT.baseIndex-blockIndex)
	for i := 0; i < len(result); i++ {
		result[i] = cobiT.blocks[cobiT.baseIndex-blockIndex-uint32(i)-1]
	}
	return result
}
