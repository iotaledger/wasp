package database

import (
	"fmt"
	"runtime"

	hivedb "github.com/iotaledger/hive.go/kvstore/database"
	"github.com/iotaledger/hive.go/kvstore/flushkv"
	"github.com/iotaledger/hive.go/kvstore/rocksdb"
)

// NewRocksDB creates a new RocksDB instance.
func NewRocksDB(path string) (*rocksdb.RocksDB, error) {
	opts := []rocksdb.Option{
		rocksdb.IncreaseParallelism(runtime.NumCPU() - 1),
		rocksdb.Custom([]string{
			"periodic_compaction_seconds=43200",
			"level_compaction_dynamic_level_bytes=true",
			"keep_log_file_num=2",
			"max_log_file_size=50000000", // 50MB per log file
		}),
	}

	return rocksdb.CreateDB(path, opts...)
}

func newDatabaseRocksDB(path string, autoFlush bool) (*Database, error) {
	rocksDatabase, err := NewRocksDB(path)
	if err != nil {
		return nil, fmt.Errorf("rocksdb database initialization failed: %w", err)
	}

	store := rocksdb.New(rocksDatabase)
	if autoFlush {
		store = flushkv.New(store)
	}

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
