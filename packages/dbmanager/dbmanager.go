package dbmanager

import (
	"fmt"
	"sync"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/dbprovider"
	"github.com/iotaledger/wasp/packages/parameters"
	"github.com/labstack/gommon/log"
	"github.com/mr-tron/base58"
	"golang.org/x/xerrors"
)

type DBManager struct {
	log         *logger.Logger
	dbInstances map[[ledgerstate.AddressLength]byte]*dbprovider.DBProvider
	mutex       *sync.RWMutex
}

func NewDBManager(log *logger.Logger) *DBManager {
	return &DBManager{
		log:         log,
		dbInstances: make(map[[ledgerstate.AddressLength]byte]*dbprovider.DBProvider),
		mutex:       &sync.RWMutex{},
	}
}

func (m *DBManager) Close() {
	for _, instance := range m.dbInstances {
		instance.Close()
	}
}

func (m *DBManager) createInstance(chainID *coretypes.ChainID) *dbprovider.DBProvider {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// if chainID is nil, no dbInstance exists for the chain records registry
	if chainID != nil {
		chainRecord, err := m.ChainRecordFromRegistry(chainID)
		if err != nil {
			panic(xerrors.Errorf("chainRecord not found for chainID: %s", chainID))
		}

		if !chainRecord.DedicatedDbInstance {
			//use a common dbProvider (same as the db provider for the chain records)
			instance := m.GetDBInstance(nil)
			m.dbInstances[chainID.Array()] = instance
			return instance
		}
	} else {
		chainID = &coretypes.ChainID{}
	}

	// create new db instance
	if parameters.GetBool(parameters.DatabaseInMemory) {
		log.Infof("IN MEMORY DATABASE")
		instance := dbprovider.NewInMemoryDBProvider(m.log)
		m.dbInstances[chainID.Array()] = instance
		return instance
	}

	dbDir := parameters.GetString(parameters.DatabaseDir)
	instanceDir := fmt.Sprintf("%s/%s", dbDir, chainID.Base58())
	instance := dbprovider.NewPersistentDBProvider(instanceDir, m.log)
	m.dbInstances[chainID.Array()] = instance
	return instance
}

func (m *DBManager) GetRegistryPartition() kvstore.KVStore {
	return m.GetKVStore(nil)
}

func DbKeyChainRecord(chainID *coretypes.ChainID) []byte {
	return MakeKey(ObjectTypeChainRecord, chainID.Bytes())
}

// ChainRecordFromRegistry reads ChainRecord from registry.
// Returns nil if not found
func (m *DBManager) ChainRecordFromRegistry(chainID *coretypes.ChainID) (*coretypes.ChainRecord, error) {
	data, err := m.GetRegistryPartition().Get(DbKeyChainRecord(chainID))
	if err == kvstore.ErrKeyNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return coretypes.ChainRecordFromBytes(data)
}

func (m *DBManager) GetDBInstance(chainID *coretypes.ChainID) *dbprovider.DBProvider {
	if chainID == nil {
		chainID = &coretypes.ChainID{} // chain records registry
	}
	return m.dbInstances[chainID.Array()]
}

func (m *DBManager) GetKVStore(chainID *coretypes.ChainID) kvstore.KVStore {
	instance := m.GetDBInstance(chainID)
	if instance != nil {
		return instance.GetPartition(chainID)
	}
	// create a new instance
	return m.createInstance(chainID).GetPartition(chainID)
}

func (m *DBManager) GetChainRecords() ([]*coretypes.ChainRecord, error) {
	db := m.GetRegistryPartition()
	ret := make([]*coretypes.ChainRecord, 0)

	err := db.Iterate([]byte{ObjectTypeChainRecord}, func(key kvstore.Key, value kvstore.Value) bool {
		if rec, err1 := coretypes.ChainRecordFromBytes(value); err1 == nil {
			ret = append(ret, rec)
		} else {
			log.Warnf("corrupted chain record with key %s", base58.Encode(key))
		}
		return true
	})
	return ret, err
}

func (m *DBManager) RunGC(shutdownSignal <-chan struct{}) {
	for _, instance := range m.dbInstances {
		instance.RunGC(shutdownSignal)
	}
}

// ------------------------------------------------------------------------

// func (rec *ChainRecord) SaveToRegistry() error {
// 	return database.GetRegistryPartition().Set(DbKeyChainRecord(rec.ChainID), rec.Bytes())
// }

// func UpdateChainRecord(chainID *ChainID, f func(*ChainRecord) bool) (*ChainRecord, error) {
// 	rec, err := ChainRecordFromRegistry(chainID)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if rec == nil {
// 		return nil, fmt.Errorf("no chain record found for chainID %s", chainID.String())
// 	}
// 	if f(rec) {
// 		err = rec.SaveToRegistry()
// 		if err != nil {
// 			return nil, err
// 		}
// 	}
// 	return rec, nil
// }

// func ActivateChainRecord(chainID *ChainID) (*ChainRecord, error) {
// 	return UpdateChainRecord(chainID, func(bd *ChainRecord) bool {
// 		if bd.Active {
// 			return false
// 		}
// 		bd.Active = true
// 		return true
// 	})
// }

// func DeactivateChainRecord(chainID *ChainID) (*ChainRecord, error) {
// 	return UpdateChainRecord(chainID, func(bd *ChainRecord) bool {
// 		if !bd.Active {
// 			return false
// 		}
// 		bd.Active = false
// 		return true
// 	})
// }
