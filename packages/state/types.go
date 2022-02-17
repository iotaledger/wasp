package state

import (
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
	StateCommitment() hashing.HashValue
	KVStoreReader() kv.KVStoreReader
	ApplyStateUpdates(...StateUpdate)
	ApplyBlock(Block) error
	ExtractBlock() (Block, error)
	Commit(blocks ...Block) error
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

const OriginStateHashBase58 = "96yCdioNdifMb8xTeHQVQ8BzDnXDbRBoYzTq7iVaymvV"

func OriginStateHash() hashing.HashValue {
	ret, err := hashing.HashValueFromBase58(OriginStateHashBase58)
	if err != nil {
		panic(err)
	}
	return ret
}
