package database

import (
	"fmt"
	"github.com/samber/lo"
	"path"
	"sync"

	"github.com/iotaledger/hive.go/kvstore"
	hivedb "github.com/iotaledger/hive.go/db"
	"github.com/iotaledger/hive.go/runtime/options"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/registry"
)

const (
	StoreVersionChainState byte = 1
	CacheSizeDefault            = 1024 * 1024 * 32
)

type ChainStateKVStoreProvider func(chainID isc.ChainID) (kvstore.KVStore, *sync.Mutex, error)

type ChainStateDatabaseManager struct {
	mutex sync.RWMutex

	// options
	engine       hivedb.Engine
	databasePath string
	cacheSize    uint64

	// databases
	databases map[isc.ChainID]*databaseWithHealthTracker
}

func WithEngine(engine hivedb.Engine) options.Option[ChainStateDatabaseManager] {
	return func(d *ChainStateDatabaseManager) {
		d.engine = engine
	}
}

func WithPath(databasePath string) options.Option[ChainStateDatabaseManager] {
	return func(d *ChainStateDatabaseManager) {
		d.databasePath = databasePath
	}
}

func WithCacheSize(cacheSize uint64) options.Option[ChainStateDatabaseManager] {
	return func(d *ChainStateDatabaseManager) {
		d.cacheSize = cacheSize
	}
}

func NewChainStateDatabaseManager(chainRecordRegistryProvider registry.ChainRecordRegistryProvider, opts ...options.Option[ChainStateDatabaseManager]) (*ChainStateDatabaseManager, error) {
	m := options.Apply(&ChainStateDatabaseManager{
		engine:       hivedb.EngineAuto,
		databasePath: "waspdb/chains/data",
		databases:    make(map[isc.ChainID]*databaseWithHealthTracker),
	}, opts)

	// load all active chain state databases
	var innerErr error
	if err := chainRecordRegistryProvider.ForEachActiveChainRecord(func(cr *registry.ChainRecord) bool {
		_, err := m.createDatabase(cr.ChainID())
		if err != nil {
			innerErr = err
			return false
		}

		return true
	}); err != nil {
		return nil, err
	}
	if innerErr != nil {
		return nil, innerErr
	}

	return m, nil
}

func (m *ChainStateDatabaseManager) chainStateKVStore(chainID isc.ChainID) (kvstore.KVStore, *sync.Mutex) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	databaseChainState, exists := m.databases[chainID]
	if !exists {
		return nil, nil
	}

	return databaseChainState.database.KVStore(), &databaseChainState.database.writeMutex
}

func (m *ChainStateDatabaseManager) createDatabase(chainID isc.ChainID) (*databaseWithHealthTracker, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if databaseChainState, exists := m.databases[chainID]; exists {
		return databaseChainState, nil
	}

	databaseChainState, err := newDatabaseWithHealthTracker(
		path.Join(m.databasePath, chainID.String()),
		m.engine,
		false,
		m.cacheSize,
		StoreVersionChainState,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("chain state database initialization failed: %w", err)
	}

	m.databases[chainID] = databaseChainState
	return databaseChainState, nil
}

func (m *ChainStateDatabaseManager) ChainStateKVStore(chainID isc.ChainID) (kvstore.KVStore, *sync.Mutex, error) {
	if store, writeMutex := m.chainStateKVStore(chainID); store != nil {
		return store, writeMutex, nil
	}

	databaseChainState, err := m.createDatabase(chainID)
	if err != nil {
		return nil, nil, err
	}

	return databaseChainState.database.KVStore(), &databaseChainState.database.writeMutex, nil
}

func (m *ChainStateDatabaseManager) FlushAndCloseStores() error {
	var err error

	// Flush all databases
	for _, db := range lo.Values(m.databases) {
		if errTmp := db.Flush(); errTmp != nil {
			err = errTmp
		}
	}

	// Close all databases
	for _, db := range lo.Values(m.databases) {
		if errTmp := db.Close(); errTmp != nil {
			err = errTmp
		}
	}

	return err
}

func (m *ChainStateDatabaseManager) MarkStoresCorrupted() error {
	var err error

	for _, db := range lo.Values(m.databases) {
		if errTmp := db.storeHealthTracker.MarkCorrupted(); errTmp != nil {
			err = errTmp
		}
	}

	return err
}

func (m *ChainStateDatabaseManager) MarkStoresTainted() error {
	var err error

	for _, db := range lo.Values(m.databases) {
		if errTmp := db.storeHealthTracker.MarkTainted(); errTmp != nil {
			err = errTmp
		}
	}

	return err
}

func (m *ChainStateDatabaseManager) MarkStoresHealthy() error {
	var err error

	for _, db := range lo.Values(m.databases) {
		if errTmp := db.storeHealthTracker.MarkHealthy(); errTmp != nil {
			err = errTmp
		}
	}

	return err
}

func (m *ChainStateDatabaseManager) AreStoresCorrupted() (bool, error) {
	for _, db := range lo.Values(m.databases) {
		corrupted, err := db.storeHealthTracker.IsCorrupted()
		if err != nil {
			return true, err
		}
		if corrupted {
			return true, nil
		}
	}

	return false, nil
}

func (m *ChainStateDatabaseManager) AreStoresTainted() (bool, error) {
	for _, db := range lo.Values(m.databases) {
		tainted, err := db.storeHealthTracker.IsTainted()
		if err != nil {
			return true, err
		}
		if tainted {
			return true, nil
		}
	}

	return false, nil
}

func (m *ChainStateDatabaseManager) CheckCorrectStoresVersion() (bool, error) {
	for _, db := range lo.Values(m.databases) {
		correct, err := db.storeHealthTracker.CheckCorrectStoreVersion()
		if err != nil {
			return false, err
		}
		if !correct {
			return false, nil
		}
	}

	return true, nil
}

// UpdateStoresVersion tries to migrate the existing data to the new store version.
func (m *ChainStateDatabaseManager) UpdateStoresVersion() (bool, error) {
	allCorrect := true
	for _, db := range lo.Values(m.databases) {
		_, err := db.storeHealthTracker.UpdateStoreVersion()
		if err != nil {
			return false, err
		}

		correct, err := db.storeHealthTracker.CheckCorrectStoreVersion()
		if err != nil {
			return false, err
		}
		if !correct {
			allCorrect = false
		}
	}

	return allCorrect, nil
}
