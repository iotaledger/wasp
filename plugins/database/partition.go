package database

import (
	"bytes"
	"github.com/iotaledger/goshimmer/dapps/valuetransfers/packages/address"
	"github.com/iotaledger/goshimmer/packages/database"
	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/plugins/config"
)

// database is structured with 34 byte long prefixes 'address' || 'object type byte
// '
const (
	ObjectTypeDBSchemaVersion byte = iota
	ObjectTypeBootupData
	ObjectTypeSCMetaData
	ObjectTypeDistributedKeyData
	ObjectTypeVariableState
	ObjectTypeStateUpdateBatch
	ObjectTypeProcessedRequestId
)

// storeInstance returns the KVStore instance.
func storeInstance() kvstore.KVStore {
	storeOnce.Do(createStore)
	return store
}

// storeRealm is a factory method for a different realm backed by the KVStore instance.
func storeRealm(realm kvstore.Realm) kvstore.KVStore {
	return storeInstance().WithRealm(realm)
}

// Partition returns store prefixed with the smart contract address
// Wasp ledger is partitioned by smart contract addresses
func GetPartition(addr *address.Address) kvstore.KVStore {
	return storeRealm(addr[:])
}

func GetRegistryPartition() kvstore.KVStore {
	var niladdr address.Address
	return GetPartition(&niladdr)
}

// MakeKey makes key within the partition. It consists to one byte for object type
// and arbitrary byte fragments concatenated together
func MakeKey(objType byte, keyBytes ...[]byte) []byte {
	var buf bytes.Buffer
	buf.WriteByte(objType)
	for _, b := range keyBytes {
		buf.Write(b)
	}
	return buf.Bytes()
}

func createStore() {
	log = logger.NewLogger(PluginName)

	var err error
	if config.Node.GetBool(CfgDatabaseInMemory) {
		log.Infof("IN MEMORY DATABASE")
		db, err = database.NewMemDB()
	} else {
		dbDir := config.Node.GetString(CfgDatabaseDir)
		db, err = database.NewDB(dbDir)
	}
	if err != nil {
		log.Fatal(err)
	}

	store = db.NewStore()
}

func KeyExistInPartition(addr *address.Address, objType byte, keyBytes ...[]byte) (bool, error) {
	exist, err := GetPartition(addr).Has(MakeKey(objType, keyBytes...))
	return exist, err
}
