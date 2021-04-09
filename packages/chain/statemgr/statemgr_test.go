package statemgr

import (
	"testing"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/dbprovider"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/stretchr/testify/require"
)

//---------------------------------------------
type chainForTestStateManager struct {
	chainID *coretypes.ChainID
}

func (cThis *chainForTestStateManager) ID() *coretypes.ChainID {
	return cThis.chainID
}

func (*chainForTestStateManager) EventRequestProcessed() *events.Event {
	return nil
}

func (*chainForTestStateManager) ReceiveMessage(interface{}) {
}

func (*chainForTestStateManager) Dismiss() {
}

//---------------------------------------------
//Tests if state manager is started and initialised correctly
func TestStateManager(t *testing.T) {
	log := testlogger.NewLogger(t)
	db := dbprovider.NewInMemoryDBProvider(log)

	chain := &chainForTestStateManager{chainID: coretypes.NewChainID(ledgerstate.NewAliasAddress([]byte(t.Name())))}

	manager := New(db, chain, nil, nil, log)
	time.Sleep(200 * time.Millisecond)
	require.Nil(t, manager.(*stateManager).solidState)
	require.True(t, len(manager.(*stateManager).blockCandidates) == 1)
}
