package database

import (
	"bytes"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/hive.go/database"
)

const (
	DBPrefixDatabaseVersion byte = iota
	DBPrefixSCMetaData
	DBPrefixKeyData
	DBPrefixSCState
)

// smart contract meta data
func GetSCMetaDataDB() (database.Database, error) {
	return Get(DBPrefixSCMetaData, GetBadgerInstance())
}

// smart contract meta data
func GetKeyDataDB() (database.Database, error) {
	return Get(DBPrefixKeyData, GetBadgerInstance())
}

// smart contract state data
func GetSCStateDB() (database.Database, error) {
	return Get(DBPrefixSCState, GetBadgerInstance())
}

// omitting first version byte in the address
func ObjKey(fragments ...[]byte) []byte {
	var buf bytes.Buffer
	for _, f := range fragments {
		buf.Write(f)
	}
	return buf.Bytes()
}

// omitting first version byte in the address
func ObjAddressKey(addr *address.Address, fragments ...[]byte) []byte {
	return ObjKey(append([][]byte{addr.Bytes()}, fragments...)...)
}

func DbKeyDKShare(addr *address.Address) []byte {
	return ObjAddressKey(addr)
}

func DbKeySCMetaData(addr *address.Address) []byte {
	return ObjAddressKey(addr)
}
