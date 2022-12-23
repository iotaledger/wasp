package database

import (
	hivedb "github.com/iotaledger/hive.go/core/database"
	"github.com/iotaledger/hive.go/core/kvstore/mapdb"
)

func newDatabaseMapDB() *Database {
	return New(
		"",
		mapdb.NewMapDB(),
		hivedb.EngineMapDB,
		false,
		nil,
	)
}
