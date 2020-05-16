package state

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
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
	Apply(Batch) (VariableState, error)
	// commit means saving variable state to sc db, making it persistent
	Commit(address *address.Address, batch Batch) error
	// return hash of the variable state. It is a root of the Merkle chain of all
	// state updates starting from the origin
	Hash() *hashing.HashValue
	// the storage of variable/value pairs
	Variables() variables.Variables
}

// State update represents update to the variable state
// it is calculated by the VM (in batches)
// State updates comes in batches, all state updates within one batch
// has same state index, state tx id and batch size. Batch index is unique in batch
// Batch is completed when it contains one state update for each index
type StateUpdate interface {
	StateIndex() uint32
	// transaction which validates the batch
	StateTransactionId() valuetransaction.ID
	SetStateTransactionId(valuetransaction.ID)
	BatchIndex() uint16
	// request which resulted in this state update
	RequestId() *sctransaction.RequestId
	// the payload of variables/values
	Variables() variables.Variables
	Hash() *hashing.HashValue
}

// Batch of state updates applicable to the variable state by applying state updates
// in a sequence defined by batch indices
type Batch interface {
	ForEach(func(StateUpdate) bool) bool
	StateIndex() uint32
	// transaction which validates the batch
	StateTransactionId() valuetransaction.ID
	SetStateTransactionId(valuetransaction.ID)
	Size() uint16
	RequestIds() []*sctransaction.RequestId
	// is a hash of all hashes of state updates. It is only correct when valid
	Hash() *hashing.HashValue
}
