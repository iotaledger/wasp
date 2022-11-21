package state

import (
	"sync"
	"time"

	"github.com/iotaledger/hive.go/core/kvstore"
	"github.com/iotaledger/hive.go/core/kvstore/mapdb"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
)

// InitChainStore initializes a new chain store, committing the origin block and state to the DB.
// The ChainID is not known at this point, so it will contain all zeroes.
func InitChainStore(db kvstore.KVStore) Store {
	store := NewStore(db)
	d := store.NewOriginStateDraft()
	emptyChainID := isc.ChainID{}
	d.Set(KeyChainID, emptyChainID.Bytes())
	d.Set(kv.Key(coreutil.StatePrefixBlockIndex), codec.EncodeUint32(0))
	d.Set(kv.Key(coreutil.StatePrefixTimestamp), codec.EncodeTime(time.Unix(0, 0)))
	block := store.Commit(d)
	err := store.SetLatest(block.TrieRoot())
	if err != nil {
		panic(err)
	}
	return store
}

// OriginL1Commitment calculates the L1Commitment for the origin block, which is
// the same for every chain.
func OriginL1Commitment() *L1Commitment {
	originL1CommitmentOnce.Do(calcOriginL1Commitment)
	return originL1Commitment
}

var (
	originL1Commitment     *L1Commitment
	originL1CommitmentOnce sync.Once
)

func calcOriginL1Commitment() {
	store := InitChainStore(mapdb.NewMapDB())
	block, err := store.LatestBlock()
	if err != nil {
		panic(err)
	}
	originL1Commitment = block.L1Commitment()
}
