package dbmanager

import (
	"testing"

	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/stretchr/testify/require"
)

func TestNewDbManager(t *testing.T) {
	log := testlogger.NewLogger(t)
	dbm := NewDBManager(log, true, registry.DefaultConfig())
	require.NotNil(t, dbm.registryDB)
	require.NotNil(t, dbm.registryStore)
	require.Empty(t, dbm.databases)
	require.Empty(t, dbm.stores)
}

func TestCreateDb(t *testing.T) {
	log := testlogger.NewLogger(t)
	dbm := NewDBManager(log, true, registry.DefaultConfig())
	chainID := iscp.RandomChainID()
	require.Nil(t, dbm.GetKVStore(chainID))
	require.NotNil(t, dbm.GetOrCreateKVStore(chainID))
	require.Len(t, dbm.databases, 1)
	require.Len(t, dbm.stores, 1)
}
