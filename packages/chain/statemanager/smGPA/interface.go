package smGPA

import (
	"github.com/iotaledger/wasp/packages/state"
)

type createStateFun func() (state.VirtualStateAccess, error)

const topPriority = uint32(0)

type blockRequest interface {
	getLastBlockHash() state.BlockHash
	getLastBlockIndex() uint32 // TODO: temporary function. Remove it after DB refactoring.
	blockAvailable(state.Block)
	isValid() bool
	getPriority() uint32
	markCompleted(createStateFun) // NOTE: not all the requests need the base state, so a function to create one is passed rather than the created state
}
