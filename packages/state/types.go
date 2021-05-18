package state

import (
	"github.com/iotaledger/wasp/packages/kv"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/buffered"
)

// StateReader read-only access to the chain's state
type StateReader interface {
	BlockIndex() uint32
	Timestamp() time.Time
	Hash() hashing.HashValue
	KVStoreReader() kv.KVStoreReader
}

// VirtualState virtualized access to the chain's state
type VirtualState interface {
	StateReader
	ApplyStateUpdates(...StateUpdate)
	ApplyBlock(Block) error
	ExtractBlock() (Block, error)
	Commit(blocks ...Block) error
	KVStore() *buffered.BufferedKVStore
	Clone() VirtualState
	DangerouslyConvertToString() string
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
	ApprovingOutputID() ledgerstate.OutputID
	SetApprovingOutputID(ledgerstate.OutputID)
	Timestamp() time.Time
	EssenceBytes() []byte // except state transaction id
	Bytes() []byte
}

const (
	OriginStateHashBase58 = "4Rx7PFaQTyyYEeESYXhjbYQNpzhzbWQM6uwePRGw3U1V"
)

func OriginStateHash() hashing.HashValue {
	ret, err := hashing.HashValueFromBase58(OriginStateHashBase58)
	if err != nil {
		panic(err)
	}
	return ret
}
