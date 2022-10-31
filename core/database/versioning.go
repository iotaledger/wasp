package database

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/iotaledger/hive.go/core/kvstore"
	"github.com/iotaledger/wasp/packages/database/dbkeys"
	"github.com/iotaledger/wasp/packages/hashing"
)

const (
	// DBVersion defines the version of the database schema this version of Wasp supports.
	// Every time there's a breaking change regarding the stored data, this version flag should be adjusted.
	DBVersion = 0
)

// ErrDBVersionIncompatible is returned when the database has an unexpected version.
var ErrDBVersionIncompatible = errors.New("database version is not compatible. please delete your database folder and restart")

// checks whether the database is compatible with the current schema version.
// also automatically sets the version if the database if new.
// version is stored in niladdr partition.
// it consists of one byte of version and the hash (checksum) of that one byte
func checkDatabaseVersion() error { //nolint:deadcode,unused
	dbRegistry := deps.DatabaseManager.GetRegistryKVStore()
	ver, err := dbRegistry.Get(dbkeys.MakeKey(dbkeys.ObjectTypeDBSchemaVersion))

	var versiondata [1 + hashing.HashSize]byte
	versiondata[0] = DBVersion
	vh := hashing.HashStrings(fmt.Sprintf("dbversion = %d", DBVersion))
	copy(versiondata[1:], vh[:])

	if errors.Is(err, kvstore.ErrKeyNotFound) {
		// set the version in an empty DB
		return dbRegistry.Set(dbkeys.MakeKey(dbkeys.ObjectTypeDBSchemaVersion), versiondata[:])
	}
	if err != nil {
		return err
	}
	if len(ver) == 0 {
		return fmt.Errorf("%w: no database version was persisted", ErrDBVersionIncompatible)
	}
	if !bytes.Equal(ver, versiondata[:]) {
		return fmt.Errorf("%w: supported version: %d, version of database: %d", ErrDBVersionIncompatible, DBVersion, ver[0])
	}
	return nil
}
