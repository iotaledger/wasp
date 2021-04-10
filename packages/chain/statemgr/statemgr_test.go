package statemgr

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate/utxodb"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/utxoutil"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/chain/mock_chain"
	"github.com/iotaledger/wasp/packages/state"
	"go.uber.org/zap/zapcore"
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
	ch := mock_chain.NewMockedChain(*chainID, log)
	nodeConn := mock_chain.NewMockedNodeConnection()
	peers := mock_chain.NewDummyPeerGroup()

	manager := New(db, ch, peers, nodeConn, log)
	time.Sleep(200 * time.Millisecond)
	require.Nil(t, manager.(*stateManager).solidState)
	require.True(t, len(manager.(*stateManager).blockCandidates) == 1)
}

func TestGetInitialState(t *testing.T) {
	log := testlogger.WithLevel(testlogger.NewLogger(t), zapcore.InfoLevel, false)
	db := dbprovider.NewInMemoryDBProvider(log)
	ledger := utxodb.New()
	keyPair, addr := ledger.NewKeyPairByIndex(0)
	_, err := ledger.RequestFunds(addr)
	require.NoError(t, err)

	origiTx, chainOutput := initChain(t, ledger, keyPair, addr)
	chainID := coretypes.NewChainID(chainOutput.GetAliasAddress())

	ch := mock_chain.NewMockedChain(*chainID, log)
	peers := mock_chain.NewDummyPeerGroup()
	nodeConn := mock_chain.NewMockedNodeConnection()
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

func TestGetNextState(t *testing.T) {
	log := testlogger.WithLevel(testlogger.NewLogger(t), zapcore.InfoLevel, false)
	db := dbprovider.NewInMemoryDBProvider(log)
	ledger := utxodb.New()
	keyPair, addr := ledger.NewKeyPairByIndex(0)
	_, err := ledger.RequestFunds(addr)
	require.NoError(t, err)

	origiTx, chainOutput := initChain(t, ledger, keyPair, addr)
	chainID := coretypes.NewChainID(chainOutput.GetAliasAddress())

	ch := mock_chain.NewMockedChain(*chainID, log)
	peers := mock_chain.NewDummyPeerGroup()
	nodeConn := mock_chain.NewMockedNodeConnection()
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
	require.EqualValues(t, 0, manager.(*stateManager).stateOutput.GetStateIndex())
	require.EqualValues(t, manager.(*stateManager).solidState.Hash(), state.OriginStateHash())

	//-------------------------------------------------------------

	currentState := manager.(*stateManager).solidState
	require.NotNil(t, currentState)
	currentStateOutput := manager.(*stateManager).stateOutput
	require.NotNil(t, currentState)
	currh := currentState.Hash()
	require.EqualValues(t, currh[:], currentStateOutput.GetStateData())

	stateTransition := mock_chain.NewMockedStateTransition(t, ledger, keyPair)
	stateTransition.OnNextState(func(block state.Block, tx *ledgerstate.Transaction) {
		log.Infof("++++++++++++ OnNextState")
		go manager.EventBlockCandidateMsg(chain.BlockCandidateMsg{Block: block})
		go func() {
			err := ledger.AddTransaction(tx)
			require.NoError(t, err)
		}()
		nextChainOut, err := utxoutil.GetSingleChainedAliasOutput(tx)
		require.NoError(t, err)
		go manager.EventStateMsg(&chain.StateMsg{
			ChainOutput: nextChainOut,
			Timestamp:   tx.Essence().Timestamp(),
		})
	})
	stateTransition.NextState(currentState, currentStateOutput)
	time.Sleep(200 * time.Millisecond)

	require.EqualValues(t, 1, manager.(*stateManager).stateOutput.GetStateIndex())
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
