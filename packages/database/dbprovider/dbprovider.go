package dbprovider

import (
	"sync"
	"time"

	"github.com/iotaledger/goshimmer/packages/database"
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/timeutil"
	"github.com/iotaledger/wasp/packages/coretypes"
)

type DBProvider struct {
	log             *logger.Logger
	db              database.DB
	store           kvstore.KVStore
	partitions      map[[ledgerstate.AddressLength]byte]kvstore.KVStore
	partitionsMutex *sync.RWMutex
}

func newDBProvider(db database.DB, log *logger.Logger) *DBProvider {
	return &DBProvider{
		log:             log,
		db:              db,
		store:           db.NewStore(),
		partitions:      make(map[[ledgerstate.AddressLength]byte]kvstore.KVStore),
		partitionsMutex: &sync.RWMutex{},
	}
}

func NewInMemoryDBProvider(log *logger.Logger) *DBProvider {
	db, err := database.NewMemDB()
	if err != nil {
		log.Fatal(err)
	}
	return newDBProvider(db, log)
}

func NewPersistentDBProvider(dbDir string, log *logger.Logger) *DBProvider {
	db, err := database.NewDB(dbDir)
	if err != nil {
		log.Fatal(err)
	}
	return newDBProvider(db, log)
}

// GetPartition returns a Partition, which is a KVStore prefixed with the chain ID.
func (dbp *DBProvider) GetPartition(chainID *coretypes.ChainID) kvstore.KVStore {
	var chainIDArr [ledgerstate.AddressLength]byte
	if chainID != nil {
		chainIDArr = chainID.Array()
	}
	dbp.partitionsMutex.RLock()
	ret, ok := dbp.partitions[chainIDArr]
	if ok {
		defer dbp.partitionsMutex.RUnlock()
		return ret
	}
	// switching to write lock
	dbp.partitionsMutex.RUnlock()
	dbp.partitionsMutex.Lock()
	defer dbp.partitionsMutex.Unlock()

	dbp.partitions[chainIDArr] = dbp.store.WithRealm(chainIDArr[:])
	return dbp.partitions[chainIDArr]
}

func (dbp *DBProvider) GetRegistryPartition() kvstore.KVStore {
	return dbp.GetPartition(nil)
}

func (dbp *DBProvider) Close() {
	dbp.log.Infof("Syncing database to disk...")
	if err := dbp.db.Close(); err != nil {
		dbp.log.Errorf("Failed to flush the database: %s", err)
	}
	dbp.log.Infof("Syncing database to disk... done")
}

func (dbp *DBProvider) RunGC(shutdownSignal <-chan struct{}) {
	if !dbp.db.RequiresGC() {
		return
	}
	// run the garbage collection with the given interval
	timeutil.NewTicker(func() {
		if err := dbp.db.GC(); err != nil {
			dbp.log.Warnf("Garbage collection failed: %s", err)
		}
	}, 5*time.Minute, shutdownSignal)
}
