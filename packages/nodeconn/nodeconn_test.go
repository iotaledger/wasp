package nodeconn

import (
	"net"
	"testing"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/utxodb"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/utxoutil"
	"github.com/iotaledger/goshimmer/packages/waspconn"
	"github.com/iotaledger/goshimmer/packages/waspconn/connector"
	"github.com/iotaledger/goshimmer/packages/waspconn/utxodbledger"
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/logger"
	"github.com/stretchr/testify/require"
)

func start(t *testing.T) (*utxodbledger.UtxoDBLedger, *NodeConn) {
	t.Helper()

	ledger := utxodbledger.New()
	t.Cleanup(ledger.Detach)

	dial := DialFunc(func() (string, net.Conn, error) {
		conn1, conn2 := net.Pipe()
		connector.Run(conn2, logger.NewExampleLogger("waspconn"), ledger)
		return "pipe", conn1, nil
	})

	n := New("test", logger.NewExampleLogger("nodeconn"), dial, "goshimerTest")
	t.Cleanup(n.Close)

	ok := n.WaitForConnection(10 * time.Second)
	require.True(t, ok)

	return ledger, n
}

func send(t *testing.T, n *NodeConn, sendMsg func() error, rcv func(msg waspconn.Message) bool) {
	t.Helper()

	done := make(chan bool)

	closure := events.NewClosure(func(msg waspconn.Message) {
		t.Logf("received msg from waspconn %T", msg)
		if !rcv(msg) {
			close(done)
		}
	})

	n.Events.MessageReceived.Attach(closure)
	defer n.Events.MessageReceived.Detach(closure)

	err := sendMsg()
	require.NoError(t, err)

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("timeout")
	}
}

func createChain(t *testing.T, u *utxodbledger.UtxoDBLedger, creatorIndex int, stateControlIndex int, balances map[ledgerstate.Color]uint64) *ledgerstate.AliasAddress {
	t.Helper()

	creatorKP, creatorAddr := utxodb.NewKeyPairByIndex(creatorIndex)
	err := u.RequestFunds(creatorAddr)
	require.NoError(t, err)

	_, addrStateControl := utxodb.NewKeyPairByIndex(stateControlIndex)
	outputs := u.GetAddressOutputs(creatorAddr)
	txb := utxoutil.NewBuilder(outputs...)
	err = txb.AddNewChainMint(balances, addrStateControl, nil)
	require.NoError(t, err)
	err = txb.AddReminderOutputIfNeeded(creatorAddr, nil)
	require.NoError(t, err)
	tx, err := txb.BuildWithED25519(creatorKP)
	require.NoError(t, err)

	err = u.PostTransaction(tx)
	require.NoError(t, err)

	chainOutput, err := utxoutil.GetSingleChainedOutput(tx.Essence())
	require.NoError(t, err)

	return chainOutput.GetAliasAddress()
}

func TestRequestBacklog(t *testing.T) {
	const (
		creatorIndex      = 2
		stateControlIndex = 3
	)

	ledger, n := start(t)

	chainAddress := createChain(t, ledger, creatorIndex, stateControlIndex, map[ledgerstate.Color]uint64{ledgerstate.ColorIOTA: 100})
	t.Logf("chain address: %s", chainAddress.Base58())

	// request backlog for chainAddress
	var resp *waspconn.WaspFromNodeTransactionMsg
	send(t, n,
		func() error {
			return n.RequestBacklogFromNode(chainAddress)
		},
		func(msg waspconn.Message) bool {
			switch msg := msg.(type) {
			case *waspconn.WaspFromNodeTransactionMsg:
				resp = msg
				return false
			}
			return true
		},
	)

	// assert response message
	require.EqualValues(t, chainAddress.Base58(), resp.ChainAddress.Base58())

	_, creatorAddr := utxodb.NewKeyPairByIndex(creatorIndex)
	t.Logf("creator address: %s", creatorAddr.Base58())

	require.EqualValues(t, creatorAddr.Base58(), resp.Sender.Base58())

	chainOutput, err := utxoutil.GetSingleChainedOutput(resp.Tx.Essence())
	require.NoError(t, err)
	require.EqualValues(t, chainAddress.Base58(), chainOutput.Address().Base58())
}

/*
func postRequest(t *testing.T, u *utxodbledger.UtxoDBLedger, fromIndex int, chainAddress *ledgerstate.AliasAddress) *ledgerstate.Transaction {
	kp, addr := utxodb.NewKeyPairByIndex(fromIndex)

	outs := u.GetAddressOutputs(addr)

	txb := utxoutil.NewBuilder(outs...)
	txb.AddExtendedOutputSimple(chainAddress, []byte{1, 3, 3, 7}, nil)
	err := txb.AddReminderOutputIfNeeded(addr, nil)
	require.NoError(t, err)
	tx, err := txb.BuildWithED25519(kp)
	require.NoError(t, err)

	err = u.PostTransaction(tx)
	require.NoError(t, err)

	return tx
}

func TestPostRequest(t *testing.T) {
	ledger, n := start(t)

	// transfer 1337 iotas to addr
	seed := ed25519.NewSeed()
	addr := ledgerstate.NewED25519Address(seed.KeyPair(0).PublicKey)
	err := ledger.RequestFunds(addr)
	require.NoError(t, err)

	// transfer 1 iota from fromAddr to addr2
	addr2 := ledgerstate.NewED25519Address(seed.KeyPair(1).PublicKey)
	tx := transfer(t, ledger, seed.KeyPair(0), addr2, 1)

	// post tx
	err = n.PostTransactionToNode(tx, addr, 0)
	require.NoError(t, err)

	// request tx
	var txMsg *waspconn.WaspFromNodeConfirmedTransactionMsg
	send(t, n, &txMsg, func() error {
		return n.RequestConfirmedTransactionFromNode(tx.ID())
	})
	require.EqualValues(t, txMsg.Tx.ID(), tx.ID())
}
*/

/*
func TestRequestInclusionLevel(t *testing.T) {
	ledger, n := start(t)

	// transfer 1337 iotas to addr
	seed := ed25519.NewSeed()
	addr := ledgerstate.NewED25519Address(seed.KeyPair(0).PublicKey)
	err := ledger.RequestFunds(addr)
	require.NoError(t, err)

	// find out tx id
	var txID ledgerstate.TransactionID
	for outID := range ledger.GetAddressOutputs(addr) {
		txID = outID.TransactionID()
	}
	require.NotEqualValues(t, ledgerstate.TransactionID{}, txID)

	// request inclusion level
	var resp *waspconn.WaspFromNodeBranchInclusionStateMsg
	send(t, n, &resp, func() error {
		return n.RequestBranchInclusionStateFromNode(txID, addr)
	})
	require.EqualValues(t, ledgerstate.Confirmed, resp.State)
}

func TestSubscribe(t *testing.T) {
	ledger, n := start(t)

	// transfer 1337 iotas to addr
	seed := ed25519.NewSeed()
	addr := ledgerstate.NewED25519Address(seed.KeyPair(0).PublicKey)
	err := ledger.RequestFunds(addr)
	require.NoError(t, err)

	// subscribe to addr
	n.Subscribe(addr, ledgerstate.ColorIOTA)
	n.log.Debugf("XXX before")
	time.Sleep(5 * time.Second)
	n.log.Debugf("XXX after")

	// transfer 1 iota from fromAddr to addr2
	addr2 := ledgerstate.NewED25519Address(seed.KeyPair(1).PublicKey)
	tx := transfer(t, ledger, seed.KeyPair(0), addr2, 1)

	// request tx
	var txMsg *waspconn.WaspFromNodeAddressUpdateMsg
	send(t, n, &txMsg, func() error {
		return n.PostTransactionToNode(tx, addr, 0)
	})
	require.EqualValues(t, txMsg.Tx.ID(), tx.ID())
}
*/
