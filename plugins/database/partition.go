package database

import (
	"bytes"
	"sync"

	"github.com/iotaledger/goshimmer/packages/database"
	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/parameters"
)

// each key in DB is prefixed with `chainID` | `SC index` | `object type byte`

const (
	ObjectTypeDBSchemaVersion byte = iota
	ObjectTypeBootupData
	ObjectTypeDistributedKeyData
	ObjectTypeSolidState
	ObjectTypeStateUpdateBatch
	ObjectTypeProcessedRequestId
	ObjectTypeSolidStateIndex
	ObjectTypeStateVariable
	ObjectTypeProgramMetadata
	ObjectTypeProgramCode
)

type Partition struct {
	kvstore.KVStore
	mut *sync.RWMutex
}

var (
	// to be able to work with MapsDB
	partitions      = make(map[coretypes.ChainID]*Partition)
	partitionsMutex sync.RWMutex
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

// GetPartition returns a Partition, which is a KVStore prefixed with the chain ID.
func GetPartition(chainID *coretypes.ChainID) *Partition {
	partitionsMutex.RLock()
	ret, ok := partitions[*chainID]
	if ok {
		defer partitionsMutex.RUnlock()
		return ret
	}
	// switching to write lock
	partitionsMutex.RUnlock()
	partitionsMutex.Lock()
	defer partitionsMutex.Unlock()

	partitions[*chainID] = &Partition{
		KVStore: storeRealm(chainID[:]),
		mut:     &sync.RWMutex{},
	}
	return partitions[*chainID]
}

func GetRegistryPartition() kvstore.KVStore {
	var niladdr coretypes.ChainID
	return GetPartition(&niladdr)
}

func (part *Partition) RLock() {
	part.mut.RLock()
}

func (part *Partition) RUnlock() {
	part.mut.RUnlock()
}

func (part *Partition) Lock() {
	part.mut.Lock()
}

func (part *Partition) Unlock() {
	part.mut.Unlock()
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
	if parameters.GetBool(parameters.DatabaseInMemory) {
		log.Infof("IN MEMORY DATABASE")
		db, err = database.NewMemDB()
	} else {
		dbDir := parameters.GetString(parameters.DatabaseDir)
		db, err = database.NewDB(dbDir)
	}
	if err != nil {
		log.Fatal(err)
	}

	store = db.NewStore()
}
