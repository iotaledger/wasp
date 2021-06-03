package state

import (
	"time"

	"github.com/iotaledger/wasp/packages/kv"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/kv/buffered"
)

// VirtualState virtualized access to the chain's database
type VirtualState interface {
	BlockIndex() uint32
	Timestamp() time.Time
	Hash() hashing.HashValue
	KVStoreReader() kv.KVStoreReader
	ApplyStateUpdates(...StateUpdate)
	ApplyBlock(Block) error
	ExtractBlock() (Block, error)
	Commit(blocks ...Block) error
	KVStore() *buffered.BufferedKVStore
	Clone() VirtualState
	DangerouslyConvertToString() string
}

type StateReader interface {
	BlockIndex() (uint32, error)
	Timestamp() (time.Time, error)
	Hash() (hashing.HashValue, error)
	KVStoreReader() kv.KVStoreReader
}

type OptimisticStateReader interface {
	StateReader
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
