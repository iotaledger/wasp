package dbmanager

import (
	"runtime"

	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/iotaledger/hive.go/kvstore/rocksdb"
)

// region Interface ////////////////////////////////////////////////////////////

// DB represents a database abstraction.
type DB interface {
	// NewStore creates a new KVStore backed by the database.
	NewStore() kvstore.KVStore
	// Close closes a DB.
	Close() error

	// RequiresGC returns whether the database requires a call of GC() to clean deleted items.
	RequiresGC() bool
	// GC runs the garbage collection to clean deleted database items.
	GC() error
}

// endregion ///////////////////////////////////////////////////////////////////

// region in-memory DB /////////////////////////////////////////////////////////

type memDB struct {
	kvstore.KVStore
}

// NewMemDB returns a new in-memory (not persisted) DB object.
func NewMemDB() (DB, error) {
	return &memDB{KVStore: mapdb.NewMapDB()}, nil
}

func (db *memDB) NewStore() kvstore.KVStore {
	return db.KVStore
}

func (db *memDB) Close() error {
	db.KVStore = nil
	return nil
}

func (db *memDB) RequiresGC() bool {
	return false
}

func (db *memDB) GC() error {
	return nil
}

// endregion ///////////////////////////////////////////////////////////////////

// region rocksDB //////////////////////////////////////////////////////////////

type rocksDB struct {
	*rocksdb.RocksDB
}

// NewDB returns a new persisting DB object.
func NewDB(dirname string) (DB, error) {
	//opt := rocksdb.UseCompression(true)
	//db, err := rocksdb.CreateDB(dirname, opt)
	db, err := rocksdb.CreateDB(dirname)
	return &rocksDB{RocksDB: db}, err
}

func (db *rocksDB) NewStore() kvstore.KVStore {
	return rocksdb.New(db.RocksDB)
}

// Close closes a DB. It's crucial to call it to ensure all the pending updates make their way to disk.
func (db *rocksDB) Close() error {
	return db.RocksDB.Close()
}

func (db *rocksDB) RequiresGC() bool {
	return true
}

func (db *rocksDB) GC() error {
	// trigger the go garbage collector to release the used memory
	runtime.GC()
	return nil
}

// endregion ///////////////////////////////////////////////////////////////////
