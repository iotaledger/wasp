package database

import (
	"fmt"
	"runtime"

	hivedb "github.com/iotaledger/hive.go/db"
	"github.com/iotaledger/wasp/v2/packages/kvstore/rocksdb"
)

// NewRocksDB creates a new RocksDB instance.
func NewRocksDB(
	path string,
	cacheSize uint64,
	bloomFilterBitsPerKey float64,
) (*rocksdb.RocksDB, error) {
	opts := []rocksdb.Option{
		rocksdb.IncreaseParallelism(runtime.NumCPU() - 1),
		rocksdb.BlockCacheSize(cacheSize),
		rocksdb.BloomFilterBitsPerKey(bloomFilterBitsPerKey),
		rocksdb.Custom([]string{
			"stats_dump_period_sec=10",
			"periodic_compaction_seconds=43200",
			"level_compaction_dynamic_level_bytes=true",
			"keep_log_file_num=2",
			"max_log_file_size=50000000", // 50MB per log file
		}),
	}

	return rocksdb.CreateDB(path, opts...)
}

func newDatabaseRocksDB(
	path string,
	cacheSize uint64,
	bloomFilterBitsPerKey float64,
) (*Database, error) {
	rocksDatabase, err := NewRocksDB(path, cacheSize, bloomFilterBitsPerKey)
	if err != nil {
		return nil, fmt.Errorf("rocksdb database initialization failed: %w", err)
	}

	store := rocksdb.New(rocksDatabase)
	return New(
		path,
		store,
		hivedb.EngineRocksDB,
		true,
		func() bool {
			if numCompactions, success := rocksDatabase.GetIntProperty("rocksdb.num-running-compactions"); success {
				running := numCompactions != 0
				return running
			}
			return false
		},
	), nil
}

func NewReadOnlyDatabase(
	path string,
) (*Database, error) {
	dbConn, err := rocksdb.OpenDBReadOnly(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open read-only RocksDB: %w", err)
	}

	db := New(path, rocksdb.New(dbConn), hivedb.EngineRocksDB, false, func() bool {
		return false
	})
	return db, nil
}
