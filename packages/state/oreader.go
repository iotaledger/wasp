package state

import (
	"time"

	"github.com/iotaledger/hive.go/core/kvstore"
	"github.com/iotaledger/trie.go/trie"
	"github.com/iotaledger/wasp/packages/common"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/optimism"
)

// optimisticStateReaderImpl state reader reads the chain state from db and validates it
type optimisticStateReaderImpl struct {
	stateReader *optimism.OptimisticKVStoreReader
	trie        trie.NodeStore
}

// NewOptimisticStateReader creates new optimistic read-only access to the database. It contains own read baseline
func NewOptimisticStateReader(store kvstore.KVStore, glb coreutil.ChainStateSync) *optimisticStateReaderImpl { //nolint:revive
	chainReader := kv.NewHiveKVStoreReader(subRealm(store, []byte{common.ObjectTypeState}))
	baseline := glb.GetSolidIndexBaseline()
	return &optimisticStateReaderImpl{
		stateReader: optimism.NewOptimisticKVStoreReader(chainReader, baseline),
		trie:        NewTrieReader(trieKVStore(store), valueKVStore(store)),
	}
}

func (r *optimisticStateReaderImpl) ChainID() (*isc.ChainID, error) {
	chidBin, err := r.stateReader.Get("")
	if err != nil {
		return nil, err
	}
	return isc.ChainIDFromBytes(chidBin)
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
