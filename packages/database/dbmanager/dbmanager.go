package dbmanager

import (
	"fmt"
	"sync"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/kvstore"
	"github.com/iotaledger/hive.go/logger"
	"github.com/iotaledger/wasp/packages/database/dbprovider"
	"github.com/iotaledger/wasp/packages/parameters"
)

type DBManager struct {
	log         *logger.Logger
	dbInstances map[[ledgerstate.AddressLength]byte]*dbprovider.DBProvider
	mutex       *sync.RWMutex
	inMemory    bool
}

var dbmanager *DBManager

// used to set the instance for testing purposes // TODO check if there is a better method
func SetInstance(dbm *DBManager) {
	dbmanager = dbm
}

func Instance() *DBManager {
	if dbmanager == nil {
		dbmanager = NewDBManager(logger.NewLogger("dbmanager"), parameters.GetBool(parameters.DatabaseInMemory))
	}
	return dbmanager
}

func NewDBManager(logger *logger.Logger, inMemory bool) *DBManager {
	return &DBManager{
		log:         logger,
		dbInstances: make(map[[ledgerstate.AddressLength]byte]*dbprovider.DBProvider),
		mutex:       &sync.RWMutex{},
		inMemory:    inMemory,
	}
}

func (m *DBManager) Close() {
	for _, instance := range m.dbInstances {
		instance.Close()
	}
}

func (m *DBManager) createDedicatedDbInstance(chainID *ledgerstate.AliasAddress) *dbprovider.DBProvider {
	// create new db instance
	if m.inMemory {
		m.log.Infof("IN MEMORY DATABASE")
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

func (m *DBManager) createInstance(chainID *ledgerstate.AliasAddress, dedicatedDbInstance bool) *dbprovider.DBProvider {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	var instance *dbprovider.DBProvider
	if dedicatedDbInstance {
		instance = m.createDedicatedDbInstance(chainID)
	} else {
		instance = m.GetRegistryDbInstance()
	}

	m.dbInstances[chainID.Array()] = instance
	return instance
}

func (m *DBManager) GetOrCreateDBInstance(chainID *ledgerstate.AliasAddress, dedicatedDbInstance bool) *dbprovider.DBProvider {
	if chainID == nil {
		// chain records registry
		return m.GetOrCreateDBInstance(&ledgerstate.AliasAddress{}, true)
	}
	instance := m.dbInstances[chainID.Array()]
	if instance == nil {
		instance = m.createInstance(chainID, dedicatedDbInstance)
	}
	return instance
}

func (m *DBManager) GetDBInstance(chainID *ledgerstate.AliasAddress) *dbprovider.DBProvider {
	return m.dbInstances[chainID.Array()]
}

func (m *DBManager) GetRegistryDbInstance() *dbprovider.DBProvider {
	zeroAddress := ledgerstate.AliasAddress{}
	instance := m.dbInstances[zeroAddress.Array()]
	if instance == nil {
		// first call, registry instance does not exist yet
		instance = m.GetOrCreateDBInstance(nil, true)
	}
	return instance
}

func (m *DBManager) GetRegistryKVStore() kvstore.KVStore {
	return m.GetRegistryDbInstance().GetPartition(nil)
}

func (m *DBManager) GetOrCreateKVStore(chainID *ledgerstate.AliasAddress, dedicatedDbInstance bool) kvstore.KVStore {
	instance := m.GetDBInstance(chainID)
	if instance != nil {
		return instance.GetPartition(chainID)
	}
	// create a new instance
	return m.GetOrCreateDBInstance(chainID, dedicatedDbInstance).GetPartition(chainID)
}

func (m *DBManager) GetKVStore(chainID *ledgerstate.AliasAddress) kvstore.KVStore {
	return m.GetDBInstance(chainID).GetPartition(chainID)
}

func (m *DBManager) RunGC(shutdownSignal <-chan struct{}) {
	for _, instance := range m.dbInstances {
		instance.RunGC(shutdownSignal)
	}
}
