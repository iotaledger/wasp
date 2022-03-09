package state

import (
	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/wasp/packages/database/dbkeys"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/optimism"
	"github.com/iotaledger/wasp/packages/kv/trie"
	"time"
)

// OptimisticStateReaderImpl state reader reads the chain state from db and validates it
type OptimisticStateReaderImpl struct {
	db          kvstore.KVStore
	stateReader *optimism.OptimisticKVStoreReader
	trie        trie.NodeStore
}

// NewOptimisticStateReader creates new optimistic read-only access to the database. It contains own read baseline
func NewOptimisticStateReader(db kvstore.KVStore, glb coreutil.ChainStateSync) *OptimisticStateReaderImpl {
	chainReader := kv.NewHiveKVStoreReader(subRealm(db, []byte{dbkeys.ObjectTypeState}))
	trieReader := kv.NewHiveKVStoreReader(subRealm(db, []byte{dbkeys.ObjectTypeTrie}))
	baseline := glb.GetSolidIndexBaseline()
	return &OptimisticStateReaderImpl{
		db:          db,
		stateReader: optimism.NewOptimisticKVStoreReader(chainReader, baseline),
		trie:        trie.NewNodeStore(optimism.NewOptimisticKVStoreReader(trieReader, baseline), CommitmentModel),
	}
}

func (r *OptimisticStateReaderImpl) BlockIndex() (uint32, error) {
	blockIndex, err := loadStateIndexFromState(r.stateReader)
	if err != nil {
		return 0, err
	}
	return blockIndex, nil
}

func (r *OptimisticStateReaderImpl) Timestamp() (time.Time, error) {
	ts, err := loadTimestampFromState(r.stateReader)
	if err != nil {
		return time.Time{}, err
	}
	return ts, nil
}

func (r *OptimisticStateReaderImpl) KVStoreReader() kv.KVStoreReader {
	return r.stateReader
}

func (r *OptimisticStateReaderImpl) SetBaseline() {
	r.stateReader.SetBaseline()
}

func (r *OptimisticStateReaderImpl) TrieNodeStore() trie.NodeStore {
	return r.trie
}
