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

func TestPostTransaction(t *testing.T) {
	tangle, n := start(t)

	// transfer 1337 iotas to addr
	seed := ed25519.NewSeed()
	addr := ledgerstate.NewED25519Address(seed.KeyPair(0).PublicKey)
	err := tangle.RequestFunds(addr)
	require.NoError(t, err)

	// request utxos for addr
	var addrOutsMsg *waspconn.WaspFromNodeAddressOutputsMsg
	doAndWaitForResponse(t, n, &addrOutsMsg, func() error {
		return n.RequestOutputsFromNode(addr)
	})
	var outID ledgerstate.OutputID
	for txID := range addrOutsMsg.Balances {
		var txMsg *waspconn.WaspFromNodeConfirmedTransactionMsg
		doAndWaitForResponse(t, n, &txMsg, func() error {
			return n.RequestConfirmedTransactionFromNode(txID)
		})
		for _, out := range txMsg.Tx.Essence().Outputs() {
			if out.Address().Array() != addr.Array() {
				continue
			}
			require.EqualValues(t, 1337, out.Balances().Map()[ledgerstate.ColorIOTA])
			outID = out.ID()
		}
	}
	require.NotEqualValues(t, ledgerstate.EmptyOutputID, outID)

	// transfer 1 iota from addr to addr2
	addr2 := ledgerstate.NewED25519Address(seed.KeyPair(1).PublicKey)
	pledge, _ := identity.RandomID()
	txEssence := ledgerstate.NewTransactionEssence(
		0,
		time.Now(),
		pledge,
		pledge,
		ledgerstate.NewInputs(ledgerstate.NewUTXOInput(outID)),
		ledgerstate.NewOutputs(
			ledgerstate.NewSigLockedSingleOutput(uint64(1), addr2),
			ledgerstate.NewSigLockedSingleOutput(uint64(1336), addr),
		),
	)
	sig := ledgerstate.NewED25519Signature(seed.KeyPair(0).PublicKey, ed25519.Signature(seed.KeyPair(0).PrivateKey.Sign(txEssence.Bytes())))
	unlockBlock := ledgerstate.NewSignatureUnlockBlock(sig)
	tx := ledgerstate.NewTransaction(txEssence, ledgerstate.UnlockBlocks{unlockBlock})

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
