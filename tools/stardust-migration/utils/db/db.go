package db

import (
	"runtime"

	"github.com/samber/lo"

	hivedb "github.com/iotaledger/hive.go/db"
	"github.com/iotaledger/hive.go/kvstore"
	old_kvstore "github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/hive.go/kvstore/rocksdb"
	"github.com/iotaledger/wasp/packages/database"
	"github.com/iotaledger/wasp/tools/stardust-migration/utils/cli"
	old_database "github.com/nnikolash/wasp-types-exported/packages/database"
)

func Create(dbDir string) kvstore.KVStore {
	cli.Logf("Creating DB in %v\n", dbDir)

	const cacheSiize64MiB = 64 * 1024 * 1024 // 64 MiB - taken from config_defaults.json
	rocksDB := lo.Must(database.NewRocksDB(dbDir, cacheSiize64MiB))

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

func openDBConnection(dbDir string) *rocksdb.RocksDB {
	return lo.Must(rocksdb.OpenDBReadOnly(dbDir,
		rocksdb.IncreaseParallelism(runtime.NumCPU()-1),
		rocksdb.ReadFillCache(true),
		rocksdb.WriteDisableWAL(true),
		rocksdb.BlockCacheSize(40*1024*1024),
		rocksdb.Custom([]string{
			"periodic_compaction_seconds=43200",
			"level_compaction_dynamic_level_bytes=true",
			"keep_log_file_num=0",
			"max_log_file_size=50000000", // 50MB per log file
		}),
	))
}

func ConnectOld(dbDir string) old_kvstore.KVStore {
	cli.Logf("Connecting to DB in %v\n", dbDir)

	conn := openDBConnection(dbDir)

	db := old_database.New(
		dbDir,
		rocksdb.New(conn),
		hivedb.EngineRocksDB,
		true,
		func() bool { panic("should not be called") },
	)

	kvs := db.KVStore()

	return kvs
}

func ConnectNew(dbDir string) kvstore.KVStore {
	cli.Logf("Connecting to DB in %v\n", dbDir)

	conn := openDBConnection(dbDir)

	db := database.New(
		dbDir,
		rocksdb.New(conn),
		hivedb.EngineRocksDB,
		true,
		func() bool { panic("should not be called") },
	)

	kvs := db.KVStore()

	return kvs
}
