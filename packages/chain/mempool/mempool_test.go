package mempool

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/utxodb"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/utxoutil"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/stretchr/testify/require"
	"testing"
	"time"

	"github.com/iotaledger/wasp/packages/coretypes"
)

func TestMempool(t *testing.T) {
	m := New(coretypes.NewDummyBlobCache(), testlogger.NewLogger(t))
	time.Sleep(2 * time.Second)
	m.Close()
	time.Sleep(1 * time.Second)
}

func TestAddRequest(t *testing.T) {
	pool := New(coretypes.NewDummyBlobCache(), testlogger.NewLogger(t))
	require.NotNil(t, pool)

	utxo := utxodb.New()
	keyPair, addr := utxo.NewKeyPairByIndex(0)
	_, err := utxo.RequestFunds(addr)
	require.NoError(t, err)

	outputs := utxo.GetAddressOutputs(addr)
	require.True(t, len(outputs) == 1)

	_, targetAddr := utxo.NewKeyPairByIndex(1)
	txBuilder := utxoutil.NewBuilder(outputs...)
	err = txBuilder.AddExtendedOutputConsume(targetAddr, []byte{1}, map[ledgerstate.Color]uint64{ledgerstate.ColorIOTA: 1})
	require.NoError(t, err)
	err = txBuilder.AddReminderOutputIfNeeded(addr, nil)
	require.NoError(t, err)
	tx, err := txBuilder.BuildWithED25519(keyPair)
	require.NoError(t, err)
	require.NotNil(t, tx)

	requests, err := sctransaction.RequestsOnLedgerFromTransaction(tx, targetAddr)
	require.NoError(t, err)
	require.NotNil(t, requests)

	pool.ReceiveRequest(requests[0])
	require.True(t, pool.HasRequest(requests[0].ID()))
}
