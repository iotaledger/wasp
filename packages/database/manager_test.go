package database

import (
	"testing"

	"github.com/stretchr/testify/require"

	hivedb "github.com/iotaledger/hive.go/core/database"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/registry"
)

func TestNewDbManager(t *testing.T) {
	dbManager, err := NewManager(registry.NewChainRecordRegistry(nil), WithEngine(hivedb.EngineMapDB))
	require.NoError(t, err)

	require.NotNil(t, dbManager.ConsensusStateKVStore())
	require.Empty(t, dbManager.databasesChainState)
}

func TestCreateDb(t *testing.T) {
	dbManager, err := NewManager(registry.NewChainRecordRegistry(nil), WithEngine(hivedb.EngineMapDB))
	require.NoError(t, err)

	chainID := isc.RandomChainID()
	require.Nil(t, dbManager.ChainStateKVStore(chainID))
	store, err := dbManager.GetOrCreateChainStateKVStore(chainID)
	require.NoError(t, err)
	require.NotNil(t, store)
	require.Len(t, dbManager.databasesChainState, 1)
}
