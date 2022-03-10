package state

import (
	"github.com/iotaledger/wasp/packages/kv/trie"
	"time"

	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/buffered"
)

// VirtualStateAccess is a virtualized access interface to the chain's database
// It consists of state reader and the buffer to collect mutations to key values
type VirtualStateAccess interface {
	BlockIndex() uint32
	Timestamp() time.Time
	TrieAccess() trie.NodeStore
	PreviousStateCommitment() trie.VCommitment
	Commit()
	ReconcileTrie() []kv.Key
	KVStoreReader() kv.KVStoreReader
	ApplyStateUpdate(Update)
	ApplyBlock(Block) error
	ProofGeneric(key []byte) *trie.ProofGeneric
	ExtractBlock() (Block, error)
	Save(blocks ...Block) error
	KVStore() *buffered.BufferedKVStoreAccess
	Copy() VirtualStateAccess
	DangerouslyConvertToString() string
}

type OptimisticStateReader interface {
	BlockIndex() (uint32, error)
	Timestamp() (time.Time, error)
	KVStoreReader() kv.KVStoreReader
	SetBaseline()
	TrieNodeStore() trie.NodeStore
}

// Update is a set of mutations
type Update interface {
	Mutations() *buffered.Mutations
	Clone() Update
	Bytes() []byte
	String() string
}

// Block is a wrapped update
type Block interface {
	BlockIndex() uint32
	ApprovingOutputID() *iotago.UTXOInput
	SetApprovingOutputID(*iotago.UTXOInput)
	Timestamp() time.Time
	PreviousStateCommitment(trie.CommitmentModel) trie.VCommitment
	EssenceBytes() []byte // except state transaction id
	Bytes() []byte
}
