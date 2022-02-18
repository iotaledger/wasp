package state

import (
	"github.com/iotaledger/wasp/packages/kv/trie"
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/buffered"
)

// VirtualStateAccess is a virtualized access interface to the chain's database
// It consists of state reader and the buffer to collect mutations to key values
type VirtualStateAccess interface {
	BlockIndex() uint32
	Timestamp() time.Time
	PreviousStateHash() hashing.HashValue
	Commit()
	StateCommitment() trie.VectorCommitment
	KVStoreReader() kv.KVStoreReader
	ApplyStateUpdate(StateUpdate)
	ApplyBlock(Block) error
	ExtractBlock() (Block, error)
	Save(blocks ...Block) error
	KVStore() *buffered.BufferedKVStoreAccess
	Copy() VirtualStateAccess
	DangerouslyConvertToString() string
}

type OptimisticStateReader interface {
	BlockIndex() (uint32, error)
	Timestamp() (time.Time, error)
	Hash() (hashing.HashValue, error)
	KVStoreReader() kv.KVStoreReader
	SetBaseline()
}

// StateUpdate is a set of mutations
type StateUpdate interface {
	Mutations() *buffered.Mutations
	Clone() StateUpdate
	Bytes() []byte
	String() string
}

// Block is a sequence of state updates applicable to the virtual state
type Block interface {
	BlockIndex() uint32
	ApprovingOutputID() *iotago.UTXOInput
	SetApprovingOutputID(*iotago.UTXOInput)
	Timestamp() time.Time
	PreviousStateHash() hashing.HashValue
	EssenceBytes() []byte // except state transaction id
	Bytes() []byte
}
