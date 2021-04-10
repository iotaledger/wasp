package statemgr

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate/utxodb"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/utxoutil"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/chain/statemgr/test_statemgr"
	"github.com/iotaledger/wasp/packages/state"
	"testing"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/dbprovider"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/stretchr/testify/require"
)

//---------------------------------------------
//Tests if state manager is started and initialised correctly
func TestStateManager(t *testing.T) {
	log := testlogger.NewLogger(t)
	db := dbprovider.NewInMemoryDBProvider(log)
	chainID := coretypes.NewChainID(ledgerstate.NewAliasAddress([]byte(t.Name())))
	chain := test_statemgr.NewMockedChain(*chainID, log)
	nodeConn := test_statemgr.NewMockedNodeConnection()
	peers := test_statemgr.NewDummyPeerGroup()

	manager := New(db, chain, peers, nodeConn, log)
	time.Sleep(200 * time.Millisecond)
	require.Nil(t, manager.(*stateManager).solidState)
	require.True(t, len(manager.(*stateManager).blockCandidates) == 1)
}

func TestGetInitialState(t *testing.T) {
	log := testlogger.NewLogger(t)
	db := dbprovider.NewInMemoryDBProvider(log)
	ledger := utxodb.New()
	keyPair, addr := ledger.NewKeyPairByIndex(0)
	_, err := ledger.RequestFunds(addr)
	require.NoError(t, err)

	origiTx, chainOutput := initChain(t, ledger, keyPair, addr)
	chainID := coretypes.NewChainID(chainOutput.GetAliasAddress())

	ch := test_statemgr.NewMockedChain(*chainID, log)
	peers := test_statemgr.NewDummyPeerGroup()
	nodeConn := test_statemgr.NewMockedNodeConnection()
	manager := New(db, ch, peers, nodeConn, log)

	nodeConn.OnPullBacklog(func(addr ledgerstate.Address) {
		log.Infof("pulled backlog addr: %s", addr)
		go manager.EventStateMsg(&chain.StateMsg{
			ChainOutput: chainOutput,
			Timestamp:   origiTx.Essence().Timestamp(),
		})
	})
	manager.EventTimerMsg(2)

	time.Sleep(200 * time.Millisecond)
	require.True(t, chainOutput.Compare(manager.(*stateManager).stateOutput) == 0)
	require.True(t, manager.(*stateManager).stateOutput.GetStateIndex() == 0)
	require.EqualValues(t, manager.(*stateManager).solidState.Hash(), state.OriginStateHash())
}

func initChain(t *testing.T, ledger *utxodb.UtxoDB, keyPair *ed25519.KeyPair, statAddr ledgerstate.Address) (*ledgerstate.Transaction, *ledgerstate.AliasOutput) {
	addr := ledgerstate.NewED25519Address(keyPair.PublicKey)
	outputs := ledger.GetAddressOutputs(addr)
	require.True(t, len(outputs) == 1)

	bals := map[ledgerstate.Color]uint64{ledgerstate.ColorIOTA: 100}
	txBuilder := utxoutil.NewBuilder(outputs...)
	err := txBuilder.AddNewAliasMint(bals, statAddr, state.OriginStateHash().Bytes())
	require.NoError(t, err)
	err = txBuilder.AddReminderOutputIfNeeded(addr, nil)
	require.NoError(t, err)
	tx, err := txBuilder.BuildWithED25519(keyPair)
	require.NoError(t, err)
	err = ledger.AddTransaction(tx)
	require.NoError(t, err)

	ret, err := utxoutil.GetSingleChainedAliasOutput(tx)
	require.NoError(t, err)

	return tx, ret
}
