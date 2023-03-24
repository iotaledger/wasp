package smGPA

import (
	"github.com/iotaledger/wasp/packages/state"
)

type (
	obtainStateFun func() (state.State, error)
	obtainBlockFun func(*state.L1Commitment) state.Block
	blockRequestID uint64
)

type blockRequest interface {
	getLastL1Commitment() *state.L1Commitment
	isValid() bool
	commitmentAvailable(*state.L1Commitment)
	getCommitmentChain() []*state.L1Commitment // NOTE: block commitments are returned in decreasing index order
	getChainOfBlocks(uint32, obtainBlockFun) chainOfBlocks
	markCompleted(obtainStateFun) // NOTE: not all the requests need state, so a function to obtain one is passed rather than the created state
	getType() string
	getID() blockRequestID
}

type chainOfBlocks interface {
	getL1Commitment(blockIndex uint32) *state.L1Commitment
	getBlocksFrom(blockIndex uint32) []state.Block // Not including blockIndex block; In proper order
}

type blockRequestCallback interface {
	isValid() bool
	requestCompleted()
}

type blockFetcher interface {
	getCommitment() *state.L1Commitment
	notifyFetched(func(blockFetcher) (bool, error)) error // calls fun for this fetcher and each related recursively; fun for parent block is always called before fun for related block
	addCallback(blockRequestCallback)
	addRelatedFetcher(blockFetcher)
	cleanCallbacks()
}

type blockFetchers interface {
	getSize() int
	addFetcher(blockFetcher)
	takeFetcher(*state.L1Commitment) blockFetcher
	addCallback(*state.L1Commitment, blockRequestCallback) bool
	addRelatedFetcher(*state.L1Commitment, blockFetcher) bool
	getCommitments() []*state.L1Commitment
	cleanCallbacks()
}
