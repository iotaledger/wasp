package statemgr

import (
	"testing"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/utxodb"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/utxoutil"
	"github.com/iotaledger/hive.go/events"
	chain_pkg "github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/dbprovider"
	"github.com/iotaledger/wasp/packages/hashing"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/stretchr/testify/require"
)

//---------------------------------------------
type chainForTestGetInitialState struct {
	chainID *coretypes.ChainID
}

func (cThis *chainForTestGetInitialState) ID() *coretypes.ChainID {
	return cThis.chainID
}

func (*chainForTestGetInitialState) EventRequestProcessed() *events.Event {
	return nil
}

func (*chainForTestGetInitialState) ReceiveMessage(interface{}) {
}

func (*chainForTestGetInitialState) Dismiss() {
}

//---------------------------------------------
type peerGroupProviderForTestGetInitialState struct{}

func (*peerGroupProviderForTestGetInitialState) NumPeers() uint16 {
	return 5
}

func (*peerGroupProviderForTestGetInitialState) NumIsAlive(quorum uint16) bool {
	return true
}

func (*peerGroupProviderForTestGetInitialState) SendMsg(targetPeerIndex uint16, msgType byte, msgData []byte) error {
	return nil
}

func (pgpThis *peerGroupProviderForTestGetInitialState) SendToAllUntilFirstError(msgType byte, msgData []byte) uint16 {
	return pgpThis.NumPeers()
}

//---------------------------------------------
type nodeConnForTestGetInitialState struct {
	inited  bool
	t       *testing.T
	tx      *ledgerstate.Transaction
	manager chain_pkg.StateManager
}

func (ncThis *nodeConnForTestGetInitialState) init() {
	if !(ncThis.inited) {
		utxo := utxodb.New()
		keyPair, addr := utxo.NewKeyPairByIndex(0)
		_, err := utxo.RequestFunds(addr)
		require.NoError(ncThis.t, err)

		outputs := utxo.GetAddressOutputs(addr)
		require.True(ncThis.t, len(outputs) == 1)

		_, addrStateControl := utxo.NewKeyPairByIndex(1)
		bals := map[ledgerstate.Color]uint64{ledgerstate.ColorIOTA: 100}
		txBuilder := utxoutil.NewBuilder(outputs...)
		err = txBuilder.AddNewAliasMint(bals, addrStateControl, hashing.RandomHash(nil).Bytes())
		require.NoError(ncThis.t, err)
		err = txBuilder.AddReminderOutputIfNeeded(addr, nil)
		require.NoError(ncThis.t, err)
		ncThis.tx, err = txBuilder.BuildWithED25519(keyPair)
		require.NoError(ncThis.t, err)
		require.NotNil(ncThis.t, ncThis.tx)
		err = utxo.AddTransaction(ncThis.tx)
		require.NoError(ncThis.t, err)

		ncThis.inited = true
	}
}

func (ncThis *nodeConnForTestGetInitialState) RequestBacklog(addr ledgerstate.Address) {
	ncThis.t.Logf("NodeConn RequestBacklog called\n")
	ncThis.init()

	aliasOutput, err := utxoutil.GetSingleChainedAliasOutput(ncThis.tx)
	require.NoError(ncThis.t, err)
	require.NotNil(ncThis.t, aliasOutput)

	eventStateMessage := &chain_pkg.StateMsg{
		ChainOutput: aliasOutput,
		Timestamp:   ncThis.tx.Essence().Timestamp(),
	}
	go ncThis.manager.EventStateMsg(eventStateMessage)
}

func (*nodeConnForTestGetInitialState) RequestConfirmedOutput(addr ledgerstate.Address, outputID ledgerstate.OutputID) {
}

func (ncThis *nodeConnForTestGetInitialState) SetManager(manager chain_pkg.StateManager) {
	ncThis.manager = manager
}

//---------------------------------------------
//Tests if started new node with empty db correctly receives initial state
func TestGetInitialState(t *testing.T) {
	log := testlogger.NewLogger(t)
	db := dbprovider.NewInMemoryDBProvider(log)

	chain := &chainForTestGetInitialState{chainID: coretypes.NewChainID(ledgerstate.NewAliasAddress([]byte(t.Name())))}
	pgp := &peerGroupProviderForTestGetInitialState{}
	nodeConn := &nodeConnForTestGetInitialState{inited: false, t: t}

	manager := New(db, chain, pgp, nodeConn, log)
	nodeConn.SetManager(manager)
	manager.EventTimerMsg(2)

	time.Sleep(200 * time.Millisecond)
	require.NotNil(t, manager.(*stateManager).stateOutput)
	require.True(t, manager.(*stateManager).stateOutput.GetStateIndex() == 0)
	require.Nil(t, manager.(*stateManager).solidState)
}
