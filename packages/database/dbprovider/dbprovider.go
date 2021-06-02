package dbprovider

import (
	"time"

	"github.com/iotaledger/goshimmer/packages/database"
	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/hive.go/timeutil"
)

type DBProvider struct {
	log   *logger.Logger
	db    database.DB
	store kvstore.KVStore
}

func newDBProvider(db database.DB, log *logger.Logger) *DBProvider {
	return &DBProvider{
		log:   log,
		db:    db,
		store: db.NewStore(),
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

func (dbp *DBProvider) GetKVStore() kvstore.KVStore {
	return dbp.store
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
