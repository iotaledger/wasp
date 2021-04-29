package state

import (
	"github.com/iotaledger/wasp/packages/kv"
	"io"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/buffered"
)

type StateReader interface {
	BlockIndex() uint32
	Timestamp() time.Time
	Hash() hashing.HashValue
	KVStoreReader() kv.KVStoreReader
}

// represents an interface to the mutable state of the smart contract
type VirtualState interface {
	StateReader
	ApplyStateUpdate(stateUpd StateUpdate)
	ApplyBlock(Block) error
	CommitToDb(block Block) error
	KVStore() *buffered.BufferedKVStore
	Clone() VirtualState
	DangerouslyConvertToString() string
}

// StateUpdate is a set of mutations
type StateUpdate interface {
	TimestampMutation() (time.Time, bool)
	StateIndexMutation() (uint32, bool)
	Hash() hashing.HashValue
	String() string
	Mutations() *buffered.Mutations
	Clone() StateUpdate
	Write(io.Writer) error
	Read(io.Reader) error
}

// Block is a sequence of state updates applicable to the variable state
type Block interface {
	ForEach(func(uint16, StateUpdate) bool)
	BlockIndex() uint32
	ApprovingOutputID() ledgerstate.OutputID
	WithApprovingOutputID(ledgerstate.OutputID) Block
	Timestamp() time.Time
	Size() uint16
	EssenceHash() hashing.HashValue // except state transaction id
	IsApprovedBy(*ledgerstate.AliasOutput) bool
	String() string
	Bytes() []byte
	Write(io.Writer) error
	Read(io.Reader) error
}
