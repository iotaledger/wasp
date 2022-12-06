package database

import (
	"fmt"
	"path"
	"sync"

	hivedb "github.com/iotaledger/hive.go/core/database"
	"github.com/iotaledger/hive.go/core/generics/lo"
	"github.com/iotaledger/hive.go/core/generics/options"
	"github.com/iotaledger/hive.go/core/kvstore"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/registry"
)

const (
	StoreVersionConsensusState byte = 1
	StoreVersionChainState     byte = 1
)

type ChainStateKVStoreProvider func(chainID isc.ChainID) (kvstore.KVStore, error)

type Manager struct {
	mutex sync.RWMutex

	// options
	engine                     hivedb.Engine
	databasePathConsensusState string
	databasesPathChainState    string

	// databases
	databaseConsensusState *databaseWithHealthTracker
	databasesChainState    map[isc.ChainID]*databaseWithHealthTracker
}

func WithEngine(engine hivedb.Engine) options.Option[Manager] {
	return func(d *Manager) {
		d.engine = engine
	}
}

func WithDatabasePathConsensusState(databasePathConsensusState string) options.Option[Manager] {
	return func(d *Manager) {
		d.databasePathConsensusState = databasePathConsensusState
	}
}

func WithDatabasesPathChainState(databasesPathChainState string) options.Option[Manager] {
	return func(d *Manager) {
		d.databasesPathChainState = databasesPathChainState
	}
}

func NewManager(chainRecordRegistryProvider registry.ChainRecordRegistryProvider, opts ...options.Option[Manager]) (*Manager, error) {
	m := options.Apply(&Manager{
		engine:                     hivedb.EngineAuto,
		databasePathConsensusState: "waspdb/chains/consensus",
		databasesPathChainState:    "waspdb/chains/data",

		databaseConsensusState: nil,
		databasesChainState:    make(map[isc.ChainID]*databaseWithHealthTracker),
	}, opts)

	// load consensus state database
	databaseConsensusState, err := newDatabaseWithHealthTracker(m.databasePathConsensusState, m.engine, true, StoreVersionConsensusState, nil)
	if err != nil {
		return nil, fmt.Errorf("consensus internal state database initialization failed: %w", err)
	}
	m.databaseConsensusState = databaseConsensusState

	// load all active chain state databases
	var innerErr error
	if err := chainRecordRegistryProvider.ForEachActiveChainRecord(func(cr *registry.ChainRecord) bool {
		_, err := m.createChainStateDatabase(cr.ChainID())
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

func (m *Manager) ConsensusStateKVStore() kvstore.KVStore {
	return m.databaseConsensusState.database.KVStore()
}

func (m *Manager) ChainStateKVStore(chainID isc.ChainID) kvstore.KVStore {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	databaseChainState, exists := m.databasesChainState[chainID]
	if !exists {
		return nil
	}

	return databaseChainState.database.KVStore()
}

func (m *Manager) createChainStateDatabase(chainID isc.ChainID) (*databaseWithHealthTracker, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if databaseChainState, exists := m.databasesChainState[chainID]; exists {
		return databaseChainState, nil
	}

	databaseChainState, err := newDatabaseWithHealthTracker(path.Join(m.databasesPathChainState, chainID.String()), m.engine, true, StoreVersionChainState, nil)
	if err != nil {
		return nil, fmt.Errorf("chain state database initialization failed: %w", err)
	}

	m.databasesChainState[chainID] = databaseChainState
	return databaseChainState, nil
}

func (m *Manager) GetOrCreateChainStateKVStore(chainID isc.ChainID) (kvstore.KVStore, error) {
	if store := m.ChainStateKVStore(chainID); store != nil {
		return store, nil
	}

	databaseChainState, err := m.createChainStateDatabase(chainID)
	if err != nil {
		return nil, err
	}

	return databaseChainState.database.KVStore(), nil
}

func (m *Manager) FlushAndCloseStores() error {
	var err error

	databases := append(lo.Values(m.databasesChainState), m.databaseConsensusState)

	// Flush all databases
	for _, db := range databases {
		if errTmp := db.Flush(); errTmp != nil {
			err = errTmp
		}
	}

	// Close all databases
	for _, db := range databases {
		if errTmp := db.Close(); errTmp != nil {
			err = errTmp
		}
	}

	return err
}

func (m *Manager) MarkStoresCorrupted() error {
	var err error

	databases := append(lo.Values(m.databasesChainState), m.databaseConsensusState)

	for _, db := range databases {
		if errTmp := db.storeHealthTracker.MarkCorrupted(); errTmp != nil {
			err = errTmp
		}
	}

	return err
}

func (m *Manager) MarkStoresTainted() error {
	var err error

	databases := append(lo.Values(m.databasesChainState), m.databaseConsensusState)

	for _, db := range databases {
		if errTmp := db.storeHealthTracker.MarkTainted(); errTmp != nil {
			err = errTmp
		}
	}

	return err
}

func (m *Manager) MarkStoresHealthy() error {
	var err error

	databases := append(lo.Values(m.databasesChainState), m.databaseConsensusState)

	for _, db := range databases {
		if errTmp := db.storeHealthTracker.MarkHealthy(); errTmp != nil {
			err = errTmp
		}
	}

	return err
}

func (m *Manager) AreStoresCorrupted() (bool, error) {
	databases := append(lo.Values(m.databasesChainState), m.databaseConsensusState)

	for _, db := range databases {
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

func (m *Manager) AreStoresTainted() (bool, error) {
	databases := append(lo.Values(m.databasesChainState), m.databaseConsensusState)

	for _, db := range databases {
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

func (m *Manager) CheckCorrectStoresVersion() (bool, error) {
	databases := append(lo.Values(m.databasesChainState), m.databaseConsensusState)

	for _, db := range databases {
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
func (m *Manager) UpdateStoresVersion() (bool, error) {
	databases := append(lo.Values(m.databasesChainState), m.databaseConsensusState)

	allCorrect := true
	for _, db := range databases {
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
