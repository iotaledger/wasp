package registry

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/core/generics/event"
	"github.com/iotaledger/wasp/packages/isc"
)

func TestNewChainStateDatabaseManager(t *testing.T) {
	chainRecordRegistry, err := NewChainRecordRegistryImpl("")
	require.NoError(t, err)

	chainID := isc.RandomChainID()

	err = chainRecordRegistry.AddChainRecord(NewChainRecord(chainID, false))
	require.NoError(t, err)

	modified := false
	active := false
	chainRecordModified := func(ev *ChainRecordModifiedEvent) {
		modified = true
		active = ev.ChainRecord.Active
	}

	chainRecordRegistry.Events().ChainRecordModified.Hook(event.NewClosure(chainRecordModified))

	_, err = chainRecordRegistry.ActivateChainRecord(chainID)
	require.NoError(t, err)

	require.True(t, modified)
	require.True(t, active)
}
