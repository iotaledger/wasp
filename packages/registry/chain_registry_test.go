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

	err = chainRecordRegistry.SaveChainRecord(registry.NewChainRecord(chainID, false, nil))
	require.NoError(t, err)

	modified := false
	active := false
	chainRecordModified := func(ev *registry.ChainRecordModifiedEvent) {
		modified = true
		active = ev.ChainRecord.Active
	}

	unhook := chainRecordRegistry.Events().ChainRecordModified.Hook(chainRecordModified).Unhook
	defer unhook()

	_, err = chainRecordRegistry.ActivateChainRecord(chainID)
	require.NoError(t, err)

	require.True(t, modified)
	require.True(t, active)
}
