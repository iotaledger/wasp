package state

import (
	"github.com/iotaledger/wasp/packages/kv"
	"io"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/buffered"
)

type StateReader interface {
	BlockIndex() uint32
	Timestamp() int64
	Hash() hashing.HashValue
	KVStoreReader() kv.KVStoreReader
}

// represents an interface to the mutable state of the smart contract
type VirtualState interface {
	StateReader
	ApplyBlockIndex(uint32)
	ApplyStateUpdate(stateUpd StateUpdate)
	ApplyBlock(Block) error
	CommitToDb(block Block) error
	KVStore() *buffered.BufferedKVStore
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
	RequestID() coretypes.RequestID
	Timestamp() int64
	WithTimestamp(int64) StateUpdate
	// the payload of variables/values
	String() string
	Mutations() *buffered.Mutations
	Clone() StateUpdate
	Write(io.Writer) error
	Read(io.Reader) error
}

// Block is a sequence of state updates applicable to the variable state
type Block interface {
	ForEach(func(uint16, StateUpdate) bool)
	StateIndex() uint32
	WithBlockIndex(uint32) Block
	ApprovingOutputID() ledgerstate.OutputID
	WithApprovingOutputID(ledgerstate.OutputID) Block
	Timestamp() int64
	Size() uint16
	RequestIDs() []coretypes.RequestID
	EssenceHash() hashing.HashValue // except state transaction id
	IsApprovedBy(*ledgerstate.AliasOutput) bool
	String() string
	Bytes() []byte
	Write(io.Writer) error
	Read(io.Reader) error
}
