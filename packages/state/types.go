package state

import (
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/variables"
	"io"
)

// represents an interface to the mutable state of the smart contract
type VariableState interface {
	// index of the current state. State index is incremented when state transition occurs
	// index 0 means origin state
	StateIndex() uint32
	// increases state index
	IncStateIndex()
	// updates state without changing state index
	ApplyStateUpdate(stateUpd StateUpdate)
	// applies batch of state updates and increases state index. Validates batch state index
	ApplyBatch(Batch) error
	// commit means saving variable state to sc db, making it persistent
	CommitToDb(address address.Address, batch Batch) error
	// return hash of the variable state. It is a root of the Merkle chain of all
	// state updates starting from the origin
	Hash() hashing.HashValue
	// the storage of variable/value pairs
	Variables() variables.Variables
}

// State update represents update to the variable state
// it is calculated by the VM (in batches)
// State updates comes in batches, all state updates within one batch
// has same state index, state tx id and batch size. ResultBatch index is unique in batch
// ResultBatch is completed when it contains one state update for each index
type StateUpdate interface {
	BatchIndex() uint16
	SetBatchIndex(uint16)
	// request which resulted in this state update
	RequestId() *sctransaction.RequestId
	// the payload of variables/values
	Variables() variables.Variables
	Write(io.Writer) error
	Read(io.Reader) error
}

// ResultBatch of state updates applicable to the variable state by applying state updates
// in a sequence defined by batch indices
type Batch interface {
	ForEach(func(StateUpdate) bool) bool
	StateIndex() uint32
	// transaction which validates the batch
	StateTransactionId() valuetransaction.ID
	Commit(valuetransaction.ID) // sets the state transaction ID
	Size() uint16
	RequestIds() []*sctransaction.RequestId
	EssenceHash() *hashing.HashValue // except state transaction id
	Write(io.Writer) error
	Read(io.Reader) error
}
