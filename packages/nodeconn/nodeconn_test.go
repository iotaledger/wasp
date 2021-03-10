package nodeconn

import (
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/goshimmer/packages/waspconn"
	"github.com/iotaledger/goshimmer/packages/waspconn/connector"
	"github.com/iotaledger/goshimmer/packages/waspconn/utxodb"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/hive.go/events"
	"github.com/iotaledger/hive.go/identity"
	"github.com/iotaledger/hive.go/logger"
	"github.com/stretchr/testify/require"
)

func start(t *testing.T) (waspconn.ValueTangle, *NodeConn) {
	t.Helper()

	tangle := utxodb.NewConfirmEmulator(0, false, false)
	t.Cleanup(tangle.Detach)

	dial := DialFunc(func() (string, net.Conn, error) {
		conn1, conn2 := net.Pipe()
		connector.Run(conn2, logger.NewExampleLogger("waspconn"), tangle)
		return "pipe", conn1, nil
	})

	n := New("test", logger.NewExampleLogger("nodeconn"), dial)
	t.Cleanup(n.Close)

	ok := n.WaitForConnection(10 * time.Second)
	require.True(t, ok)

	return tangle, n
}

func doAndWaitForResponse(t *testing.T, n *NodeConn, val interface{}, send func() error) {
	t.Helper()

	done := make(chan bool)

	vt := reflect.TypeOf(val)
	closure := events.NewClosure(func(msg interface{}) {
		mt := reflect.TypeOf(msg)
		if mt.AssignableTo(vt.Elem()) {
			reflect.ValueOf(val).Elem().Set(reflect.ValueOf(msg))
			close(done)
		} else {
			t.Logf("Received unexpected message: %T (expected %T)", msg, val)
		}
	})

	n.EventMessageReceived.Attach(closure)
	defer n.EventMessageReceived.Detach(closure)

	err := send()
	require.NoError(t, err)

	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("timeout")
	}
}

func TestRequestOutputs(t *testing.T) {
	tangle, n := start(t)

	// transfer 1337 iotas to addr
	seed := ed25519.NewSeed()
	addr := ledgerstate.NewED25519Address(seed.KeyPair(0).PublicKey)
	err := tangle.RequestFunds(addr)
	require.NoError(t, err)

	// request outputs for addr
	var msg *waspconn.WaspFromNodeAddressOutputsMsg
	doAndWaitForResponse(t, n, &msg, func() error {
		return n.RequestOutputsFromNode(addr)
	})

	// assert response message
	require.EqualValues(t, addr, msg.Address)
	require.EqualValues(t, 1, len(msg.Balances))
	for _, cb := range msg.Balances {
		cb := cb.Map()
		require.EqualValues(t, 1, len(cb))
		require.EqualValues(t, 1337, cb[ledgerstate.ColorIOTA])
	}
}

func transfer(t *testing.T, tangle waspconn.ValueTangle, from *ed25519.KeyPair, toAddr ledgerstate.Address, transferAmount uint64) *ledgerstate.Transaction {
	fromAddr := ledgerstate.NewED25519Address(from.PublicKey)

	// find an unspent output with balance >= transferAmount
	var outID ledgerstate.OutputID
	var outBalance uint64
	{
		outs := tangle.GetAddressOutputs(fromAddr)
		for id, bals := range outs {
			outBalance = bals.Map()[ledgerstate.ColorIOTA]
			outID = id
			if outBalance >= transferAmount {
				break
			}
		}
		require.GreaterOrEqual(t, outBalance, transferAmount)
	}

	pledge, _ := identity.RandomID()
	outputs := []ledgerstate.Output{
		ledgerstate.NewSigLockedSingleOutput(uint64(transferAmount), toAddr),
	}
	if outBalance > transferAmount {
		outputs = append(outputs, ledgerstate.NewSigLockedSingleOutput(uint64(outBalance-transferAmount), fromAddr))
	}
	txEssence := ledgerstate.NewTransactionEssence(
		0,
		time.Now(),
		pledge,
		pledge,
		ledgerstate.NewInputs(ledgerstate.NewUTXOInput(outID)),
		ledgerstate.NewOutputs(outputs...),
	)
	sig := ledgerstate.NewED25519Signature(from.PublicKey, ed25519.Signature(from.PrivateKey.Sign(txEssence.Bytes())))
	unlockBlock := ledgerstate.NewSignatureUnlockBlock(sig)
	return ledgerstate.NewTransaction(txEssence, ledgerstate.UnlockBlocks{unlockBlock})
}

func TestPostTransaction(t *testing.T) {
	tangle, n := start(t)

	// transfer 1337 iotas to addr
	seed := ed25519.NewSeed()
	addr := ledgerstate.NewED25519Address(seed.KeyPair(0).PublicKey)
	err := tangle.RequestFunds(addr)
	require.NoError(t, err)

	// transfer 1 iota from fromAddr to addr2
	addr2 := ledgerstate.NewED25519Address(seed.KeyPair(1).PublicKey)
	tx := transfer(t, tangle, seed.KeyPair(0), addr2, 1)

	// post tx
	err = n.PostTransactionToNode(tx, addr, 0)
	require.NoError(t, err)

	// request tx
	var txMsg *waspconn.WaspFromNodeConfirmedTransactionMsg
	doAndWaitForResponse(t, n, &txMsg, func() error {
		return n.RequestConfirmedTransactionFromNode(tx.ID())
	})
	require.EqualValues(t, txMsg.Tx.ID(), tx.ID())
}
