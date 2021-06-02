package dbmanager

import (
	"fmt"
	"sync"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/database/dbprovider"
	"github.com/iotaledger/wasp/packages/parameters"
)

type DBManager struct {
	log         *logger.Logger
	registryDB  *dbprovider.DBProvider
	dbInstances map[[ledgerstate.AddressLength]byte]*dbprovider.DBProvider
	mutex       sync.RWMutex
	inMemory    bool
}

func NewDBManager(logger *logger.Logger, inMemory bool) *DBManager {
	dbm := DBManager{
		log:         logger,
		dbInstances: make(map[[ledgerstate.AddressLength]byte]*dbprovider.DBProvider),
		mutex:       sync.RWMutex{},
		inMemory:    inMemory,
	}
	//registry db is created with an empty chainID
	dbm.registryDB = dbm.createDBInstance(&coretypes.ChainID{
		AliasAddress: &ledgerstate.AliasAddress{},
	})
	return &dbm
}

func (m *DBManager) createDBInstance(chainID *coretypes.ChainID) (instance *dbprovider.DBProvider) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.inMemory {
		m.log.Infof("creating new In-Memory database, ChainID: %s", chainID.Base58())
		instance = dbprovider.NewInMemoryDBProvider(m.log)
	} else {
		dbDir := parameters.GetString(parameters.DatabaseDir)
		instanceDir := fmt.Sprintf("%s/%s", dbDir, chainID.Base58())
		m.log.Infof("creating new persistent database, ChainID: %s, dir: %s", chainID.Base58(), instanceDir)
		instance = dbprovider.NewPersistentDBProvider(instanceDir, m.log)
	}

	m.dbInstances[chainID.Array()] = instance
	return instance
}

func (m *DBManager) getDBInstance(chainID *coretypes.ChainID) *dbprovider.DBProvider {
	return m.dbInstances[chainID.Array()]
}

func (m *DBManager) GetRegistryKVStore() kvstore.KVStore {
	return m.registryDB.GetKVStore()
}

func (m *DBManager) GetOrCreateKVStore(chainID *coretypes.ChainID) kvstore.KVStore {
	instance := m.getDBInstance(chainID)
	if instance != nil {
		return instance.GetKVStore()
	}
	// create a new instance
	return m.createDBInstance(chainID).GetKVStore()
}

func (m *DBManager) GetKVStore(chainID *coretypes.ChainID) kvstore.KVStore {
	return m.getDBInstance(chainID).GetKVStore()
}

func (m *DBManager) Close() {
	m.registryDB.Close()
	for _, instance := range m.dbInstances {
		instance.Close()
	}
}

func (m *DBManager) RunGC(shutdownSignal <-chan struct{}) {
	m.registryDB.RunGC(shutdownSignal)
	for _, instance := range m.dbInstances {
		instance.RunGC(shutdownSignal)
	}
}
