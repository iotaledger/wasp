package database

import (
	hivedb "github.com/iotaledger/hive.go/db"
	"github.com/iotaledger/wasp/packages/kvstore/mapdb"
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
