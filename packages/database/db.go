package database

import "github.com/iotaledger/hive.go/database"

func GetDB() (database.Database, error) {
	return Get(DBPrefixSmartContractLedger, database.GetBadgerInstance())
}
