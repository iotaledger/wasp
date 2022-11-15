package smGPA

import (
	"github.com/iotaledger/wasp/packages/state"
)

type createStateFun func() (state.VirtualStateAccess, error)

type blockRequestID uint64

const topPriority = uint64(0)

type blockRequest interface {
	getLastBlockHash() state.BlockHash
	getLastBlockIndex() uint32 // TODO: temporary function. Remove it after DB refactoring.
	blockAvailable(state.Block)
	isValid() bool
	getPriority() uint64
	markCompleted(createStateFun) // NOTE: not all the requests need the base state, so a function to create one is passed rather than the created state
	getType() string
	getID() blockRequestID
}
