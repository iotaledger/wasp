package registry

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/isc"
)

func TestNewChainStateDatabaseManager(t *testing.T) {
	chainRecordRegistry, err := NewChainRecordRegistryImpl("")
	require.NoError(t, err)

	chainID := isc.RandomChainID()

	err = chainRecordRegistry.AddChainRecord(NewChainRecord(chainID, false, nil))
	require.NoError(t, err)

	modified := false
	active := false
	chainRecordModified := func(ev *ChainRecordModifiedEvent) {
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
