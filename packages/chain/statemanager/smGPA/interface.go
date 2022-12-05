package smGPA

import (
	"github.com/iotaledger/wasp/packages/state"
)

type obtainStateFun func() (state.State, error)
type obtainBlockFun func(*state.L1Commitment) state.Block
type blockRequestID uint64

type blockRequest interface {
	getLastL1Commitment() *state.L1Commitment
	isValid() bool
	blockAvailable(state.Block)
	getBlockChain() []state.Block // NOTE: blocks are returned in decreasing index order
	getChainOfBlocks(uint32, obtainBlockFun) chainOfBlocks
	markCompleted(obtainStateFun) // NOTE: not all the requests need state, so a function to obtain one is passed rather than the created state
	getType() string
	getID() blockRequestID
}

type chainOfBlocks interface {
	getL1Commitment(blockIndex uint32) *state.L1Commitment
	getBlocksFrom(blockIndex uint32) []state.Block // Not including blockIndex block; In proper order
}
