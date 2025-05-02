// Package database provides interfaces and implementations for persistent data storage.
package database

import (
	"errors"
	"fmt"
	"sync"

	hivedb "github.com/iotaledger/hive.go/db"
	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/hive.go/runtime/ioutils"
	"github.com/iotaledger/wasp/packages/chaindb"
)

var AllowedEngines = []hivedb.Engine{
	hivedb.EngineMapDB,
	hivedb.EngineRocksDB,
}

type StoreVersionUpdateFunc func(store kvstore.KVStore, oldVersion byte, newVersion byte) error

// Database holds the underlying KVStore and database specific functions.
type Database struct {
	databaseDir           string
	store                 kvstore.KVStore
	engine                hivedb.Engine
	compactionSupported   bool
	compactionRunningFunc func() bool
	writeMutex            sync.Mutex
}

// New creates a new Database instance.
func New(databaseDirectory string, kvStore kvstore.KVStore, engine hivedb.Engine, compactionSupported bool, compactionRunningFunc func() bool) *Database {
	return &Database{
		databaseDir:           databaseDirectory,
		store:                 kvStore,
		engine:                engine,
		compactionSupported:   compactionSupported,
		compactionRunningFunc: compactionRunningFunc,
	}
}

// KVStore returns the underlying KVStore.
func (db *Database) KVStore() kvstore.KVStore {
	return db.store
}

// Engine returns the database engine.
func (db *Database) Engine() hivedb.Engine {
	return db.engine
}

// CompactionSupported returns whether the database engine supports compaction.
func (db *Database) CompactionSupported() bool {
	return db.compactionSupported
}

// CompactionRunning returns whether a compaction is running.
func (db *Database) CompactionRunning() bool {
	if db.compactionRunningFunc == nil {
		return false
	}

	return db.compactionRunningFunc()
}

// Size returns the size of the database.
func (db *Database) Size() (int64, error) {
	if db.engine == hivedb.EngineMapDB {
		// in-memory database does not support this method.
		return 0, nil
	}

	return ioutils.FolderSize(db.databaseDir)
}

// CheckEngine is a wrapper around hivedb.CheckEngine to throw a custom error message in case of engine mismatch.
func CheckEngine(dbPath string, createDatabaseIfNotExists bool, dbEngine hivedb.Engine) (hivedb.Engine, error) {
	targetEngine, err := hivedb.CheckEngine(dbPath, createDatabaseIfNotExists, dbEngine, AllowedEngines)
	if err != nil {
		if errors.Is(err, hivedb.ErrEngineMismatch) {
			//nolint:stylecheck // this error message is shown to the user
			return hivedb.EngineUnknown, fmt.Errorf("database (%s) engine does not match the configuration: '%v' != '%v'", dbPath, targetEngine, dbEngine[0])
		}
		return hivedb.EngineUnknown, err
	}
	return targetEngine, nil
}

func NewDatabaseInMemory() (*Database, error) {
	return NewDatabase(hivedb.EngineMapDB, "", false, false, 0)
}

// NewDatabase opens a database.
// It also checks if the database engine is correct.
func NewDatabase(
	dbEngine hivedb.Engine,
	path string,
	createDatabaseIfNotExists bool,
	autoFlush bool,
	cacheSize uint64,
) (*Database, error) {
	targetEngine, err := CheckEngine(path, createDatabaseIfNotExists, dbEngine)
	if err != nil {
		return nil, err
	}

	switch targetEngine {
	case hivedb.EngineRocksDB:
		return newDatabaseRocksDB(path, autoFlush, cacheSize)

	case hivedb.EngineMapDB:
		return newDatabaseMapDB(), nil

	default:
		return nil, fmt.Errorf("unknown database engine: %s, supported engines: rocksdb/mapdb", dbEngine)
	}
}

type databaseWithHealthTracker struct {
	database           *Database
	storeHealthTracker *kvstore.StoreHealthTracker
}

func newDatabaseWithHealthTracker(
	path string,
	dbEngine hivedb.Engine,
	autoFlush bool,
	cacheSize uint64,
	storeVersion byte,
	storeVersionUpdateFunc StoreVersionUpdateFunc,
) (*databaseWithHealthTracker, error) {
	db, err := NewDatabase(dbEngine, path, true, autoFlush, cacheSize)
	if err != nil {
		return nil, err
	}

	var hiveStoreVersionUpdateFunc kvstore.StoreVersionUpdateFunc
	if storeVersionUpdateFunc != nil {
		hiveStoreVersionUpdateFunc = func(oldVersion, newVersion byte) error {
			return storeVersionUpdateFunc(db.KVStore(), oldVersion, newVersion)
		}
	}

	healthTracker, err := kvstore.NewStoreHealthTracker(db.KVStore(), []byte{chaindb.PrefixHealthTracker}, storeVersion, hiveStoreVersionUpdateFunc)
	if err != nil {
		return nil, err
	}

	return &databaseWithHealthTracker{
		database:           db,
		storeHealthTracker: healthTracker,
	}, nil
}

func (db *databaseWithHealthTracker) Flush() error {
	var err error
	if errTmp := db.database.KVStore().Flush(); errTmp != nil {
		err = errTmp
	}

	if errTmp := db.storeHealthTracker.Flush(); errTmp != nil {
		err = errTmp
	}

	return err
}

func (db *databaseWithHealthTracker) Close() error {
	var err error
	if errTmp := db.database.KVStore().Close(); errTmp != nil {
		err = errTmp
	}

	if errTmp := db.storeHealthTracker.Close(); errTmp != nil {
		err = errTmp
	}

	return err
}
