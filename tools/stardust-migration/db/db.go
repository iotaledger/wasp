package db

import (
	"runtime"

	"github.com/iotaledger/hive.go/kvstore"
	old_kvstore "github.com/iotaledger/hive.go/kvstore"
	hivedb "github.com/iotaledger/hive.go/kvstore/database"
	"github.com/iotaledger/hive.go/kvstore/rocksdb"
	"github.com/iotaledger/wasp/packages/database"
	"github.com/iotaledger/wasp/tools/stardust-migration/cli"
	old_database "github.com/nnikolash/wasp-types-exported/packages/database"
	"github.com/samber/lo"
)

func Create(dbDir string) kvstore.KVStore {
	cli.Logf("Creating DB in %v\n", dbDir)

	// TODO: BlockCacheSize - what value should be there?
	rocksDB := lo.Must(database.NewRocksDB(dbDir, database.CacheSizeDefault))

	db := database.New(
		dbDir,
		rocksdb.New(rocksDB),
		hivedb.EngineRocksDB,
		true,
		func() bool { panic("should not be called") },
	)

	kvs := db.KVStore()

	return kvs
}

func Connect(dbDir string) old_kvstore.KVStore {
	cli.Logf("Connecting to DB in %v\n", dbDir)

	rocksDatabase := lo.Must(rocksdb.OpenDBReadOnly(dbDir,
		rocksdb.IncreaseParallelism(runtime.NumCPU()-1),
		rocksdb.Custom([]string{
			"periodic_compaction_seconds=43200",
			"level_compaction_dynamic_level_bytes=true",
			"keep_log_file_num=2",
			"max_log_file_size=50000000", // 50MB per log file
		}),
	))

	db := old_database.New(
		dbDir,
		rocksdb.New(rocksDatabase),
		hivedb.EngineRocksDB,
		true,
		func() bool { panic("should not be called") },
	)

	kvs := db.KVStore()

	return kvs
}
