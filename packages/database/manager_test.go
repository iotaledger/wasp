package database

import (
	"testing"

	"github.com/stretchr/testify/require"

	hivedb "github.com/iotaledger/hive.go/kvstore/database"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/registry"
)

func TestNewChainStateDatabaseManager(t *testing.T) {
	chainRecordRegistry, err := registry.NewChainRecordRegistryImpl("")
	require.NoError(t, err)

	chainStateDatabaseManager, err := NewChainStateDatabaseManager(chainRecordRegistry, WithEngine(hivedb.EngineMapDB))
	require.NoError(t, err)

	require.Empty(t, chainStateDatabaseManager.databases)
}

func TestCreateChainStateDatabase(t *testing.T) {
	chainRecordRegistry, err := registry.NewChainRecordRegistryImpl("")
	require.NoError(t, err)

	chainStateDatabaseManager, err := NewChainStateDatabaseManager(chainRecordRegistry, WithEngine(hivedb.EngineMapDB))
	require.NoError(t, err)

	chainID := isc.RandomChainID()
	require.Nil(t, chainStateDatabaseManager.chainStateKVStore(chainID))
	store, err := chainStateDatabaseManager.ChainStateKVStore(chainID)
	require.NoError(t, err)
	require.NotNil(t, store)
	require.Len(t, chainStateDatabaseManager.databases, 1)
}
