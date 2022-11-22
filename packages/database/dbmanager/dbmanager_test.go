package dbmanager

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/registry"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
)

func TestNewDbManager(t *testing.T) {
	log := testlogger.NewLogger(t)
	dbm := NewDBManager(log, true, "", registry.NewChainRecordRegistry(nil))
	require.Empty(t, dbm.databases)
	require.Empty(t, dbm.stores)
}

func TestCreateDb(t *testing.T) {
	log := testlogger.NewLogger(t)
	dbm := NewDBManager(log, true, "", registry.NewChainRecordRegistry(nil))
	chainID := isc.RandomChainID()
	require.Nil(t, dbm.GetKVStore(chainID))
	require.NotNil(t, dbm.GetOrCreateKVStore(chainID))
	require.Len(t, dbm.databases, 1)
	require.Len(t, dbm.stores, 1)
}
