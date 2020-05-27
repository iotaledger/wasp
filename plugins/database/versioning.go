package database

import (
	"errors"
	"fmt"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"

	"github.com/iotaledger/hive.go/kvstore"
)

const (
	// DBVersion defines the version of the database schema this version of Wasp supports.
	// Every time there's a breaking change regarding the stored data, this version flag should be adjusted.
	DBVersion = 0
)

var (
	// ErrDBVersionIncompatible is returned when the database has an unexpected version.
	ErrDBVersionIncompatible = errors.New("database version is not compatible. please delete your database folder and restart")
)

// checks whether the database is compatible with the current schema version.
// also automatically sets the version if the database is new.
// version is stored in niladdr partition
func checkDatabaseVersion() error {
	var niladdr address.Address
	db := GetPartition(&niladdr)
	ver, err := db.Get(MakeKey(ObjectTypeDBSchemaVersion))
	if err == kvstore.ErrKeyNotFound {
		// set the version in an empty DB
		return db.Set(MakeKey(ObjectTypeDBSchemaVersion), []byte{DBVersion})
	}
	if err != nil {
		return err
	}
	if len(ver) == 0 {
		return fmt.Errorf("%w: no database version was persisted", ErrDBVersionIncompatible)
	}
	if ver[0] != DBVersion {
		return fmt.Errorf("%w: supported version: %d, version of database: %d", ErrDBVersionIncompatible, DBVersion, ver[0])
	}
	return nil
}
