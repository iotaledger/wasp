package state

import (
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/variables"
)

// represents an interface to the mutable state of the smart contract
type VariableState interface {
	// index of the current state. State index is incremented when state transition occurs
	// index 0 means origin state
	StateIndex() uint32
	// state transition occurs when a batch of state updates is applied to the variable state
	// state transition produces a new transient variable state object.
	Apply([]StateUpdate) VariableState
	// commit means saving variable state to db, making it persistent
	Commit() error
	// return hash of the variable state. It is a root of the Merkle chain of all
	// state updates starting from the origin
	Hash() *hashing.HashValue

	Variables() variables.Variables
}

// State update represents update to the variable state
// it is calculated by the VM in batches
type StateUpdate interface {
	RequestId() *sctransaction.RequestId
	Variables() variables.Variables
}

// Batch of state updates applicable to the variable state
type Batch interface {
	StateIndex() uint32
	StateUpdates() []StateUpdate
	StateTransactionId() valuetransaction.ID
}
