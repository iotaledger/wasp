package mempool

import (
	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/utxodb"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/utxoutil"
	"github.com/iotaledger/wasp/packages/chain"
	"github.com/iotaledger/wasp/packages/sctransaction"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/packages/util"
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

func getRequestsOnLedger(t *testing.T, amount int) []*sctransaction.RequestOnLedger {
	utxo := utxodb.New()
	keyPair, addr := utxo.NewKeyPairByIndex(0)
	_, err := utxo.RequestFunds(addr)
	require.NoError(t, err)

	outputs := utxo.GetAddressOutputs(addr)
	require.True(t, len(outputs) == 1)

	_, targetAddr := utxo.NewKeyPairByIndex(1)
	txBuilder := utxoutil.NewBuilder(outputs...)
	var i uint64
	for i = 0; int(i) < amount; i++ {
		err = txBuilder.AddExtendedOutputConsume(targetAddr, util.Uint64To8Bytes(i), map[ledgerstate.Color]uint64{ledgerstate.ColorIOTA: 1})
		require.NoError(t, err)
	}
	err = txBuilder.AddReminderOutputIfNeeded(addr, nil)
	require.NoError(t, err)
	tx, err := txBuilder.BuildWithED25519(keyPair)
	require.NoError(t, err)
	require.NotNil(t, tx)

	requests, err := sctransaction.RequestsOnLedgerFromTransaction(tx, targetAddr)
	require.NoError(t, err)
	require.True(t, amount == len(requests))
	return requests
}

func TestAddRequest(t *testing.T) {
	pool := New(coretypes.NewDummyBlobCache(), testlogger.NewLogger(t))
	require.NotNil(t, pool)
	requests := getRequestsOnLedger(t, 1)

	pool.ReceiveRequest(requests[0])
	require.True(t, pool.HasRequest(requests[0].ID()))
}

func TestAddRemoveRequests(t *testing.T) {
	pool := New(coretypes.NewDummyBlobCache(), testlogger.NewLogger(t))
	require.NotNil(t, pool)
	requests := getRequestsOnLedger(t, 6)

	pool.ReceiveRequest(requests[0])
	pool.ReceiveRequest(requests[1])
	pool.ReceiveRequest(requests[2])
	pool.ReceiveRequest(requests[3])
	pool.ReceiveRequest(requests[4])
	pool.ReceiveRequest(requests[5])
	require.True(t, pool.HasRequest(requests[0].ID()))
	require.True(t, pool.HasRequest(requests[1].ID()))
	require.True(t, pool.HasRequest(requests[2].ID()))
	require.True(t, pool.HasRequest(requests[3].ID()))
	require.True(t, pool.HasRequest(requests[4].ID()))
	require.True(t, pool.HasRequest(requests[5].ID()))
	pool.RemoveRequests(
		requests[3].ID(),
		requests[0].ID(),
		requests[5].ID(),
	)
	require.False(t, pool.HasRequest(requests[0].ID()))
	require.True(t, pool.HasRequest(requests[1].ID()))
	require.True(t, pool.HasRequest(requests[2].ID()))
	require.False(t, pool.HasRequest(requests[3].ID()))
	require.True(t, pool.HasRequest(requests[4].ID()))
	require.False(t, pool.HasRequest(requests[5].ID()))
}

func TestTakeAllReady(t *testing.T) {
	pool := New(coretypes.NewDummyBlobCache(), testlogger.NewLogger(t))
	require.NotNil(t, pool)
	requests := getRequestsOnLedger(t, 5)

	pool.ReceiveRequest(requests[0])
	pool.ReceiveRequest(requests[1])
	pool.ReceiveRequest(requests[2])
	pool.ReceiveRequest(requests[3])
	pool.ReceiveRequest(requests[4])
	ready, result := pool.TakeAllReady(time.Now(),
		requests[0].ID(),
		requests[1].ID(),
		requests[2].ID(),
		requests[3].ID(),
		requests[4].ID(),
	)
	//NOTE: solidifaction waits for 200 miliseconds after start before proceeding
	//see packages/chain/mempool/mempool.go funcition solidificationLoop()
	require.False(t, result)
	require.Nil(t, ready)
	time.Sleep(200 * time.Millisecond)
	ready, result = pool.TakeAllReady(time.Now(),
		requests[0].ID(),
		requests[1].ID(),
		requests[2].ID(),
		requests[3].ID(),
		requests[4].ID(),
	)
	require.True(t, result)
	require.True(t, len(ready) == 5)
	require.Contains(t, ready, requests[0])
	require.Contains(t, ready, requests[1])
	require.Contains(t, ready, requests[2])
	require.Contains(t, ready, requests[3])
	require.Contains(t, ready, requests[4])
}

// Initialises the following situation
// CommiteePeer ->  0   1   2   3
//        Request0  +   +   +   +
//        Request1      +   +
//        Request2  +       +   +
//        Request3  +
//        Request4
func initSeenTest(t *testing.T) (chain.Mempool, []*sctransaction.RequestOnLedger) {
	pool := New(coretypes.NewDummyBlobCache(), testlogger.NewLogger(t))
	require.NotNil(t, pool)
	requests := getRequestsOnLedger(t, 5)
	request0ID := requests[0].ID()
	request1ID := requests[1].ID()
	request2ID := requests[2].ID()
	request3ID := requests[3].ID()

	pool.ReceiveRequest(requests[0])
	pool.MarkSeenByCommitteePeer(&request0ID, 0)

	pool.ReceiveRequest(requests[1])
	pool.MarkSeenByCommitteePeer(&request0ID, 0)
	pool.MarkSeenByCommitteePeer(&request0ID, 1)
	pool.MarkSeenByCommitteePeer(&request1ID, 1)

	pool.ReceiveRequest(requests[2])
	pool.MarkSeenByCommitteePeer(&request2ID, 0)
	pool.MarkSeenByCommitteePeer(&request2ID, 2)
	pool.MarkSeenByCommitteePeer(&request2ID, 3)

	pool.ReceiveRequest(requests[3])
	pool.MarkSeenByCommitteePeer(&request0ID, 2)
	pool.MarkSeenByCommitteePeer(&request0ID, 3)
	pool.MarkSeenByCommitteePeer(&request1ID, 2)
	pool.MarkSeenByCommitteePeer(&request3ID, 0)

	pool.ReceiveRequest(requests[4])

	return pool, requests
}

func TestGetReadyList(t *testing.T) {
	pool, requests := initSeenTest(t)

	//NOTE: solidifaction waits for 200 miliseconds after start before proceeding
	//see packages/chain/mempool/mempool.go funcition solidificationLoop()
	ready := pool.GetReadyList(0)
	require.True(t, len(ready) == 0)
	time.Sleep(200 * time.Millisecond)
	ready = pool.GetReadyList(0)
	require.True(t, len(ready) == 5)
	require.Contains(t, ready, requests[0])
	require.Contains(t, ready, requests[1])
	require.Contains(t, ready, requests[2])
	require.Contains(t, ready, requests[3])
	require.Contains(t, ready, requests[4])
	ready = pool.GetReadyList(1)
	require.True(t, len(ready) == 4)
	require.Contains(t, ready, requests[0])
	require.Contains(t, ready, requests[1])
	require.Contains(t, ready, requests[2])
	require.Contains(t, ready, requests[3])
	ready = pool.GetReadyList(2)
	require.True(t, len(ready) == 3)
	require.Contains(t, ready, requests[0])
	require.Contains(t, ready, requests[1])
	require.Contains(t, ready, requests[2])
	ready = pool.GetReadyList(3)
	require.True(t, len(ready) == 2)
	require.Contains(t, ready, requests[0])
	require.Contains(t, ready, requests[2])
	ready = pool.GetReadyList(4)
	require.True(t, len(ready) == 1)
	require.Contains(t, ready, requests[0])
	ready = pool.GetReadyList(5)
	require.True(t, len(ready) == 0)

	pool.ClearSeenMarks()
	ready = pool.GetReadyList(1)
	require.True(t, len(ready) == 0)
}

func TestGetReadyListFull(t *testing.T) {
	pool, requests := initSeenTest(t)
	request0Full := &chain.ReadyListRecord{
		Request: requests[0],
		Seen:    map[uint16]bool{0: true, 1: true, 2: true, 3: true},
	}
	request1Full := &chain.ReadyListRecord{
		Request: requests[1],
		Seen:    map[uint16]bool{1: true, 2: true},
	}
	request2Full := &chain.ReadyListRecord{
		Request: requests[2],
		Seen:    map[uint16]bool{0: true, 2: true, 3: true},
	}
	request3Full := &chain.ReadyListRecord{
		Request: requests[3],
		Seen:    map[uint16]bool{0: true},
	}
	request4Full := &chain.ReadyListRecord{
		Request: requests[4],
		Seen:    map[uint16]bool{},
	}

	//NOTE: solidifaction waits for 200 miliseconds after start before proceeding
	//see packages/chain/mempool/mempool.go funcition solidificationLoop()
	ready := pool.GetReadyListFull(0)
	require.True(t, len(ready) == 0)
	time.Sleep(200 * time.Millisecond)
	ready = pool.GetReadyListFull(0)
	require.True(t, len(ready) == 5)
	require.Contains(t, ready, request0Full)
	require.Contains(t, ready, request1Full)
	require.Contains(t, ready, request2Full)
	require.Contains(t, ready, request3Full)
	require.Contains(t, ready, request4Full)
	ready = pool.GetReadyListFull(1)
	require.True(t, len(ready) == 4)
	require.Contains(t, ready, request0Full)
	require.Contains(t, ready, request1Full)
	require.Contains(t, ready, request2Full)
	require.Contains(t, ready, request3Full)
	ready = pool.GetReadyListFull(2)
	require.True(t, len(ready) == 3)
	require.Contains(t, ready, request0Full)
	require.Contains(t, ready, request1Full)
	require.Contains(t, ready, request2Full)
	ready = pool.GetReadyListFull(3)
	require.True(t, len(ready) == 2)
	require.Contains(t, ready, request0Full)
	require.Contains(t, ready, request2Full)
	ready = pool.GetReadyListFull(4)
	require.True(t, len(ready) == 1)
	require.Contains(t, ready, request0Full)
	ready = pool.GetReadyListFull(5)
	require.True(t, len(ready) == 0)

	pool.ClearSeenMarks()
	ready = pool.GetReadyListFull(1)
	require.True(t, len(ready) == 0)
}
