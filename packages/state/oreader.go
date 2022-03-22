package state

import (
	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/wasp/packages/database/dbkeys"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/optimism"
	"github.com/iotaledger/wasp/packages/kv/trie"
	"time"
)

// optimisticStateReaderImpl state reader reads the chain state from db and validates it
type optimisticStateReaderImpl struct {
	stateReader *optimism.OptimisticKVStoreReader
	trie        trie.NodeStore
}

// NewOptimisticStateReader creates new optimistic read-only access to the database. It contains own read baseline
func NewOptimisticStateReader(db kvstore.KVStore, glb coreutil.ChainStateSync) *optimisticStateReaderImpl {
	chainReader := kv.NewHiveKVStoreReader(subRealm(db, []byte{dbkeys.ObjectTypeState}))
	trieReader := kv.NewHiveKVStoreReader(subRealm(db, []byte{dbkeys.ObjectTypeTrie}))
	baseline := glb.GetSolidIndexBaseline()
	return &optimisticStateReaderImpl{
		stateReader: optimism.NewOptimisticKVStoreReader(chainReader, baseline),
		trie:        trie.NewNodeStore(optimism.NewOptimisticKVStoreReader(trieReader, baseline), CommitmentModel),
	}
}

func (r *optimisticStateReaderImpl) ChainID() (*iscp.ChainID, error) {
	chidBin, err := r.stateReader.Get("")
	if err != nil {
		return nil, err
	}
	return iscp.ChainIDFromBytes(chidBin)
}

func (r *optimisticStateReaderImpl) BlockIndex() (uint32, error) {
	blockIndex, err := loadStateIndexFromState(r.stateReader)
	if err != nil {
		return 0, err
	}
	return blockIndex, nil
}

func (r *optimisticStateReaderImpl) Timestamp() (time.Time, error) {
	ts, err := loadTimestampFromState(r.stateReader)
	if err != nil {
		return time.Time{}, err
	}
	return ts, nil
}

func (r *optimisticStateReaderImpl) KVStoreReader() kv.KVStoreReader {
	return r.stateReader
}

func (r *optimisticStateReaderImpl) SetBaseline() {
	r.stateReader.SetBaseline()
}

func (r *optimisticStateReaderImpl) TrieNodeStore() trie.NodeStore {
	return r.trie
}
