package registry_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/packages/isc/isctest"
	"github.com/iotaledger/wasp/v2/packages/registry"
)

func TestNewChainStateDatabaseManager(t *testing.T) {
	chainRecordRegistry, err := registry.NewChainRecordRegistryImpl("")
	require.NoError(t, err)

	chainID := isctest.RandomChainID()

	err = chainRecordRegistry.SetChainRecord(registry.NewChainRecord(chainID, false, nil))
	require.NoError(t, err)

	modified := false
	active := false
	chainRecordModified := func(ev *registry.ChainRecordModifiedEvent) {
		modified = true
		active = ev.ChainRecord.Active
	}

	unhook := chainRecordRegistry.Events().ChainRecordModified.Hook(chainRecordModified).Unhook
	defer unhook()

	rec, err := chainRecordRegistry.ActivateChainRecord()
	require.NoError(t, err)
	require.NotNil(t, rec)
	require.Equal(t, chainID, rec.ChainID())

	require.True(t, modified)
	require.True(t, active)
}
