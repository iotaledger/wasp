package smGPA

import (
	"github.com/iotaledger/wasp/packages/state"
)

type obtainStateFun func() (state.State, error)

type blockRequestID uint64

const topPriority = uint64(0)

type blockRequest interface {
	getLastL1Commitment() *state.L1Commitment
	isValid() bool
	blockAvailable(state.Block)
	getBlockChain() []state.Block // NOTE: blocks are returned in decreasing index order
	markCompleted(obtainStateFun) // NOTE: not all the requests need state, so a function to obtain one is passed rather than the created state
	getType() string
	getID() blockRequestID
}
