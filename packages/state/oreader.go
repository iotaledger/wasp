package state

import (
	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/wasp/packages/database/dbkeys"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/optimism"
	"time"
)

// OptimisticStateReaderImpl state reader reads the chain state from db and validates it
type OptimisticStateReaderImpl struct {
	db         kvstore.KVStore
	chainState *optimism.OptimisticKVStoreReader
}

// NewOptimisticStateReader creates new optimistic read-only access to the database. It contains own read baseline
func NewOptimisticStateReader(db kvstore.KVStore, glb coreutil.ChainStateSync) *OptimisticStateReaderImpl {
	chainState := kv.NewHiveKVStoreReader(subRealm(db, []byte{dbkeys.ObjectTypeState}))
	return &OptimisticStateReaderImpl{
		db:         db,
		chainState: optimism.NewOptimisticKVStoreReader(chainState, glb.GetSolidIndexBaseline()),
	}
}

func (r *OptimisticStateReaderImpl) BlockIndex() (uint32, error) {
	blockIndex, err := loadStateIndexFromState(r.chainState)
	if err != nil {
		return 0, err
	}
	return blockIndex, nil
}

func (r *OptimisticStateReaderImpl) Timestamp() (time.Time, error) {
	ts, err := loadTimestampFromState(r.chainState)
	if err != nil {
		return time.Time{}, err
	}
	return ts, nil
}

func (r *OptimisticStateReaderImpl) KVStoreReader() kv.KVStoreReader {
	return r.chainState
}

func (r *OptimisticStateReaderImpl) SetBaseline() {
	r.chainState.SetBaseline()
}
