package state

import (
	"io"

	valuetransaction "github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/transaction"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/buffered"
)

// represents an interface to the mutable state of the smart contract
type VirtualState interface {
	// index of the current state. State index is incremented when state transition occurs
	// index 0 means origin state
	BlockIndex() uint32
	ApplyBlockIndex(uint32)
	// timestamp
	Timestamp() int64
	// updates state without changing state index
	ApplyStateUpdate(stateUpd StateUpdate)
	// applies block of state updates, state index and timestamp
	ApplyBlock(Block) error
	// commit means saving virtual state to sc db, making it persistent (solid)
	CommitToDb(batch Block) error
	// return hash of the variable state. It is a root of the Merkle chain of all
	// state updates starting from the origin
	Hash() hashing.HashValue
	// the storage of variable/value pairs
	Variables() buffered.BufferedKVStore
	Clone() VirtualState
	DangerouslyConvertToString() string
}

// State update represents update to the variable state
// it is calculated by the VM (in batches)
// State updates comes in batches, all state updates within one block
// has same state index, state tx id and block size. ResultBlock index is unique in block
// ResultBlock is completed when it contains one state update for each index
type StateUpdate interface {
	// request which resulted in this state update
	RequestID() *coretypes.RequestID
	Timestamp() int64
	WithTimestamp(int64) StateUpdate
	// the payload of variables/values
	String() string
	Mutations() buffered.MutationSequence
	Clone() StateUpdate
	Write(io.Writer) error
	Read(io.Reader) error
}

// Block is a sequence of state updates applicable to the variable state
type Block interface {
	ForEach(func(uint16, StateUpdate) bool)
	StateIndex() uint32
	WithBlockIndex(uint32) Block
	StateTransactionID() valuetransaction.ID
	WithStateTransaction(valuetransaction.ID) Block
	Timestamp() int64
	Size() uint16
	RequestIDs() []*coretypes.RequestID
	EssenceHash() hashing.HashValue // except state transaction id
	String() string
	Write(io.Writer) error
	Read(io.Reader) error
}
