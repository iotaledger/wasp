package nodeconn

import (
	"net"
	"testing"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/utxoutil"
	"github.com/iotaledger/goshimmer/packages/waspconn"
	"github.com/iotaledger/goshimmer/packages/waspconn/connector"
	"github.com/iotaledger/goshimmer/packages/waspconn/utxodbledger"
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/logger"
	"github.com/stretchr/testify/require"
)

const (
	creatorIndex      = 2
	stateControlIndex = 3
)

func start(t *testing.T) (*utxodbledger.UtxoDBLedger, *NodeConn) {
	t.Helper()

	ledger := utxodbledger.New()
	t.Cleanup(ledger.Detach)

	done := make(chan struct{})
	t.Cleanup(func() { close(done) })

	dial := DialFunc(func() (string, net.Conn, error) {
		conn1, conn2 := net.Pipe()
		go connector.Run(conn2, logger.NewExampleLogger("waspconn"), ledger, done)
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
		/*case <-time.After(10 * time.Second):
		t.Fatal("timeout")*/
	}
}

func createChain(t *testing.T, u *utxodbledger.UtxoDBLedger, creatorIndex int, stateControlIndex int, balances map[ledgerstate.Color]uint64) (*ledgerstate.Transaction, *ledgerstate.AliasAddress) {
	t.Helper()

	creatorKP, creatorAddr := u.NewKeyPairByIndex(creatorIndex)
	err := u.RequestFunds(creatorAddr)
	require.NoError(t, err)

	_, addrStateControl := u.NewKeyPairByIndex(stateControlIndex)
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
	chainAddress := chainOutput.GetAliasAddress()
	t.Logf("chain address: %s", chainAddress.Base58())

	return tx, chainAddress
}

func TestRequestBacklog(t *testing.T) {
	ledger, n := start(t)

	tx, chainAddress := createChain(t, ledger, creatorIndex, stateControlIndex, map[ledgerstate.Color]uint64{ledgerstate.ColorIOTA: 100})

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

	_, creatorAddr := ledger.NewKeyPairByIndex(creatorIndex)
	t.Logf("creator address: %s", creatorAddr.Base58())

	require.Equal(t, tx.ID(), resp.Tx.ID())

	chainOutput, err := utxoutil.GetSingleChainedOutput(resp.Tx.Essence())
	require.NoError(t, err)
	require.EqualValues(t, chainAddress.Base58(), chainOutput.Address().Base58())
}

func postRequest(t *testing.T, u *utxodbledger.UtxoDBLedger, fromIndex int, chainAddress *ledgerstate.AliasAddress) *ledgerstate.Transaction {
	kp, addr := u.NewKeyPairByIndex(fromIndex)

	outs := u.GetAddressOutputs(addr)

	txb := utxoutil.NewBuilder(outs...)
	err := txb.AddExtendedOutputConsume(chainAddress, []byte{1, 3, 3, 7}, map[ledgerstate.Color]uint64{ledgerstate.ColorIOTA: 1})
	require.NoError(t, err)
	err = txb.AddReminderOutputIfNeeded(addr, nil)
	require.NoError(t, err)
	tx, err := txb.BuildWithED25519(kp)
	require.NoError(t, err)

	err = u.PostTransaction(tx)
	require.NoError(t, err)

	return tx
}

func TestPostRequest(t *testing.T) {
	ledger, n := start(t)

	createTx, chainAddress := createChain(t, ledger, creatorIndex, stateControlIndex, map[ledgerstate.Color]uint64{ledgerstate.ColorIOTA: 100})

	reqTx := postRequest(t, ledger, 2, chainAddress)

	// request backlog for chainAddress
	seen := make(map[ledgerstate.TransactionID]bool)
	send(t, n,
		func() error {
			return n.RequestBacklogFromNode(chainAddress)
		},
		func(msg waspconn.Message) bool {
			switch msg := msg.(type) {
			case *waspconn.WaspFromNodeTransactionMsg:
				seen[msg.Tx.ID()] = true
				if len(seen) == 2 {
					return false
				}
			}
			return true
		},
	)

	require.Equal(t, 2, len(seen))
	require.True(t, seen[createTx.ID()])
	require.True(t, seen[reqTx.ID()])
}

func TestRequestInclusionLevel(t *testing.T) {
	ledger, n := start(t)
	createTx, chainAddress := createChain(t, ledger, creatorIndex, stateControlIndex, map[ledgerstate.Color]uint64{ledgerstate.ColorIOTA: 100})

	// request inclusion level
	var resp *waspconn.WaspFromNodeTxInclusionStateMsg
	send(t, n,
		func() error {
			return n.RequestTxInclusionStateFromNode(chainAddress, createTx.ID())
		},
		func(msg waspconn.Message) bool {
			if msg, ok := msg.(*waspconn.WaspFromNodeTxInclusionStateMsg); ok {
				resp = msg
				return false
			}
			return true
		},
	)

	require.EqualValues(t, ledgerstate.Confirmed, resp.State)
}

func TestSubscribe(t *testing.T) {
	ledger, n := start(t)
	_, chainAddress := createChain(t, ledger, creatorIndex, stateControlIndex, map[ledgerstate.Color]uint64{ledgerstate.ColorIOTA: 100})

	// subscribe to chain address
	n.Subscribe(chainAddress)

	// post a request to chain, expect to receive notification
	var reqTx *ledgerstate.Transaction
	var txMsg *waspconn.WaspFromNodeTransactionMsg
	send(t, n,
		func() error {
			reqTx = postRequest(t, ledger, 2, chainAddress)
			return nil
		},
		func(msg waspconn.Message) bool {
			switch msg := msg.(type) {
			case *waspconn.WaspFromNodeTransactionMsg:
				if msg.Tx.ID() == reqTx.ID() {
					txMsg = msg
					return false
				}
			}
			return true
		},
	)

	require.EqualValues(t, txMsg.Tx.ID(), reqTx.ID())
}
