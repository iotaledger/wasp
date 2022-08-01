package state

import (
	"time"

	"github.com/iotaledger/trie.go/trie"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/buffered"
)

// region mustOptimisticVirtualStateAccess ////////////////////////////////

// MustOptimisticVirtualState is a virtual state wrapper with global state baseline
// Once baseline is invalidated globally any subsequent access to the mustOptimisticVirtualStateAccess
// will lead to panic(coreutil.ErrorStateInvalidated)
type mustOptimisticVirtualStateAccess struct {
	state    VirtualStateAccess
	baseline coreutil.StateBaseline
}

// WrapMustOptimisticVirtualStateAccess wraps virtual state with state baseline in on object
// Does not copy buffers
func WrapMustOptimisticVirtualStateAccess(state VirtualStateAccess, baseline coreutil.StateBaseline) VirtualStateAccess {
	return &mustOptimisticVirtualStateAccess{
		state:    state,
		baseline: baseline,
	}
}

func (s *mustOptimisticVirtualStateAccess) ChainID() *isc.ChainID {
	s.baseline.MustValidate()
	defer s.baseline.MustValidate()

	return s.state.ChainID()
}

func (s *mustOptimisticVirtualStateAccess) BlockIndex() uint32 {
	s.baseline.MustValidate()
	defer s.baseline.MustValidate()

	return s.state.BlockIndex()
}

func (s *mustOptimisticVirtualStateAccess) Timestamp() time.Time {
	s.baseline.MustValidate()
	defer s.baseline.MustValidate()

	return s.state.Timestamp()
}

func (s *mustOptimisticVirtualStateAccess) WithOnBlockSave(fun OnBlockSaveClosure) {
	s.baseline.MustValidate()
	defer s.baseline.MustValidate()

	s.state.WithOnBlockSave(fun)
}

func (s *mustOptimisticVirtualStateAccess) PreviousL1Commitment() *L1Commitment {
	s.baseline.MustValidate()
	defer s.baseline.MustValidate()

	return s.state.PreviousL1Commitment()
}

func (s *mustOptimisticVirtualStateAccess) Commit() {
	s.baseline.MustValidate()
	defer s.baseline.MustValidate()

	s.state.Commit()
}

func (s *mustOptimisticVirtualStateAccess) TrieNodeStore() trie.NodeStore {
	s.baseline.MustValidate()
	defer s.baseline.MustValidate()

	return s.state.TrieNodeStore()
}

func (s *mustOptimisticVirtualStateAccess) ReconcileTrie() []kv.Key {
	s.baseline.MustValidate()
	defer s.baseline.MustValidate()

	return s.state.ReconcileTrie()
}

func (s *mustOptimisticVirtualStateAccess) KVStoreReader() kv.KVStoreReader {
	s.baseline.MustValidate()
	defer s.baseline.MustValidate()

	return s.state.KVStoreReader()
}

func (s *mustOptimisticVirtualStateAccess) OptimisticStateReader(glb coreutil.ChainStateSync) OptimisticStateReader {
	s.baseline.MustValidate()
	defer s.baseline.MustValidate()

	return s.OptimisticStateReader(glb)
}

func (s *mustOptimisticVirtualStateAccess) ApplyStateUpdate(upd Update) {
	s.baseline.MustValidate()
	defer s.baseline.MustValidate()

	s.state.ApplyStateUpdate(upd)
}

func (s *mustOptimisticVirtualStateAccess) ApplyBlock(block Block) error {
	s.baseline.MustValidate()
	defer s.baseline.MustValidate()

	return s.state.ApplyBlock(block)
}

func (s *mustOptimisticVirtualStateAccess) ProofGeneric(key []byte) *trie.ProofGeneric {
	s.baseline.MustValidate()
	defer s.baseline.MustValidate()

	return s.state.ProofGeneric(key)
}

func (s *mustOptimisticVirtualStateAccess) ExtractBlock() (Block, error) {
	s.baseline.MustValidate()
	defer s.baseline.MustValidate()

	return s.state.ExtractBlock()
}

func (s *mustOptimisticVirtualStateAccess) Save(blocks ...Block) error {
	s.baseline.MustValidate()
	defer s.baseline.MustValidate()

	return s.state.Save(blocks...)
}

func (s *mustOptimisticVirtualStateAccess) KVStore() *buffered.BufferedKVStoreAccess {
	s.baseline.MustValidate()
	defer s.baseline.MustValidate()

	return s.state.KVStore()
}

func (s *mustOptimisticVirtualStateAccess) Copy() VirtualStateAccess {
	s.baseline.MustValidate()
	defer s.baseline.MustValidate()

	return s.state.Copy()
}

func (s *mustOptimisticVirtualStateAccess) DangerouslyConvertToString() string {
	s.baseline.MustValidate()
	defer s.baseline.MustValidate()

	return s.state.DangerouslyConvertToString()
}
