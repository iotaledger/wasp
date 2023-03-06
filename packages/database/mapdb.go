package database

import (
	hivedb "github.com/iotaledger/hive.go/kvstore/database"
	"github.com/iotaledger/hive.go/kvstore/mapdb"
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
