package mempool

import (
	"testing"
	"time"

	"go.uber.org/zap/zapcore"

	"github.com/iotaledger/goshimmer/packages/ledgerstate"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/utxodb"
	"github.com/iotaledger/goshimmer/packages/ledgerstate/utxoutil"
	"github.com/iotaledger/hive.go/crypto/ed25519"
	"github.com/iotaledger/hive.go/kvstore/mapdb"
	"github.com/iotaledger/wasp/packages/coretypes"
	"github.com/iotaledger/wasp/packages/coretypes/coreutil"
	"github.com/iotaledger/wasp/packages/coretypes/request"
	"github.com/iotaledger/wasp/packages/coretypes/requestargs"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/subrealm"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/stretchr/testify/require"
)

func createStateReader(t *testing.T, glb coreutil.GlobalSync) (state.OptimisticStateReader, state.VirtualState) {
	store := mapdb.NewMapDB()
	vs, err := state.CreateOriginState(store, nil)
	require.NoError(t, err)
	ret := state.NewOptimisticStateReader(store, glb)
	require.NoError(t, err)
	return ret, vs
}

func getRequestsOnLedger(t *testing.T, amount int) ([]*request.RequestOnLedger, *ed25519.KeyPair) {
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
	err = txBuilder.AddRemainderOutputIfNeeded(addr, nil)
	require.NoError(t, err)
	tx, err := txBuilder.BuildWithED25519(keyPair)
	require.NoError(t, err)
	require.NotNil(t, tx)

	requests, err := request.RequestsOnLedgerFromTransaction(tx, targetAddr)
	require.NoError(t, err)
	require.True(t, amount == len(requests))
	return requests, keyPair
}

//Test if mempool is created
func TestMempool(t *testing.T) {
	log := testlogger.NewLogger(t)
	glb := coreutil.NewGlobalSync()
	rdr, _ := createStateReader(t, glb)
	pool := New(rdr, coretypes.NewInMemoryBlobCache(), log)
	require.NotNil(t, pool)
	time.Sleep(2 * time.Second)
	stats := pool.Stats()
	require.EqualValues(t, 0, stats.InPoolCounter)
	require.EqualValues(t, 0, stats.OutPoolCounter)
	require.EqualValues(t, 0, stats.TotalPool)
	require.EqualValues(t, 0, stats.Ready)
	pool.Close()
	time.Sleep(1 * time.Second)
}

//Test if single on ledger request is added to mempool
func TestAddRequest(t *testing.T) {
	log := testlogger.NewLogger(t)
	glb := coreutil.NewGlobalSync().SetSolidIndex(0)
	rdr, _ := createStateReader(t, glb)
	pool := New(rdr, coretypes.NewInMemoryBlobCache(), log)
	require.NotNil(t, pool)
	requests, _ := getRequestsOnLedger(t, 1)

	pool.ReceiveRequests(requests[0])
	require.True(t, pool.WaitRequestInPool(requests[0].ID()))
	stats := pool.Stats()
	require.EqualValues(t, 1, stats.InPoolCounter)
	require.EqualValues(t, 0, stats.OutPoolCounter)
	require.EqualValues(t, 1, stats.TotalPool)
	require.EqualValues(t, 1, stats.Ready)
}

func TestAddRequestInvalidState(t *testing.T) {
	log := testlogger.NewLogger(t)
	glb := coreutil.NewGlobalSync()
	glb.InvalidateSolidIndex()
	rdr, _ := createStateReader(t, glb)
	pool := New(rdr, coretypes.NewInMemoryBlobCache(), log)
	require.NotNil(t, pool)
	requests, _ := getRequestsOnLedger(t, 1)

	pool.ReceiveRequests(requests[0])
	require.False(t, pool.WaitRequestInPool(requests[0].ID(), 100*time.Millisecond))
	stats := pool.Stats()
	require.EqualValues(t, 0, stats.InPoolCounter)
	require.EqualValues(t, 0, stats.OutPoolCounter)
	require.EqualValues(t, 0, stats.TotalPool)
	require.EqualValues(t, 0, stats.Ready)

	glb.SetSolidIndex(1)
	require.True(t, pool.WaitRequestInPool(requests[0].ID(), 100*time.Millisecond))
	stats = pool.Stats()
	require.EqualValues(t, 1, stats.InPoolCounter)
	require.EqualValues(t, 0, stats.OutPoolCounter)
	require.EqualValues(t, 1, stats.TotalPool)
	require.EqualValues(t, 1, stats.Ready)
}

//Test if adding the same on ledger request more than once to the same mempool
//is handled correctly
func TestAddRequestTwice(t *testing.T) {
	log := testlogger.NewLogger(t)
	glb := coreutil.NewGlobalSync().SetSolidIndex(0)
	rdr, _ := createStateReader(t, glb)

	pool := New(rdr, coretypes.NewInMemoryBlobCache(), log)
	require.NotNil(t, pool)
	requests, _ := getRequestsOnLedger(t, 1)

	pool.ReceiveRequests(requests[0])
	require.True(t, pool.WaitRequestInPool(requests[0].ID(), 200*time.Millisecond))

	stats := pool.Stats()
	require.EqualValues(t, 1, stats.InPoolCounter)
	require.EqualValues(t, 0, stats.OutPoolCounter)
	require.EqualValues(t, 1, stats.TotalPool)
	require.EqualValues(t, 1, stats.Ready)

	pool.ReceiveRequests(requests[0])
	require.True(t, pool.WaitRequestInPool(requests[0].ID(), 200*time.Millisecond))

	stats = pool.Stats()
	require.EqualValues(t, 1, stats.InPoolCounter)
	require.EqualValues(t, 0, stats.OutPoolCounter)
	require.EqualValues(t, 1, stats.TotalPool)
	require.EqualValues(t, 1, stats.Ready)
}

//Test if adding off ledger requests works as expected: correctly signed ones
//are added, others are ignored
func TestAddOffLedgerRequest(t *testing.T) {
	log := testlogger.NewLogger(t)
	testlogger.WithLevel(log, zapcore.InfoLevel, false)
	glb := coreutil.NewGlobalSync().SetSolidIndex(0)
	rdr, _ := createStateReader(t, glb)
	pool := New(rdr, coretypes.NewInMemoryBlobCache(), log)
	require.NotNil(t, pool)
	onLedgerRequests, keyPair := getRequestsOnLedger(t, 2)

	offFromOnLedgerFun := func(onLedger *request.RequestOnLedger) *request.RequestOffLedger {
		contract, emptyPoint := onLedger.Target()
		return request.NewRequestOffLedger(contract, emptyPoint, onLedger.GetMetadata().Args())
	}
	offLedgerRequestUnsigned := offFromOnLedgerFun(onLedgerRequests[0])
	offLedgerRequestSigned := offFromOnLedgerFun(onLedgerRequests[1])
	offLedgerRequestSigned.Sign(keyPair)
	require.NotEqual(t, offLedgerRequestUnsigned.ID(), offLedgerRequestSigned.ID())

	pool.ReceiveRequests(offLedgerRequestUnsigned)
	require.False(t, pool.WaitRequestInPool(offLedgerRequestUnsigned.ID(), 200*time.Millisecond))
	stats := pool.Stats()
	require.EqualValues(t, 0, stats.InPoolCounter)
	require.EqualValues(t, 0, stats.OutPoolCounter)
	require.EqualValues(t, 0, stats.TotalPool)
	require.EqualValues(t, 0, stats.Ready)

	pool.ReceiveRequests(offLedgerRequestSigned)
	require.True(t, pool.WaitRequestInPool(offLedgerRequestSigned.ID(), 200*time.Millisecond))
	stats = pool.Stats()
	require.EqualValues(t, 1, stats.InPoolCounter)
	require.EqualValues(t, 0, stats.OutPoolCounter)
	require.EqualValues(t, 1, stats.TotalPool)
	require.EqualValues(t, 1, stats.Ready)
}

//Test if processed request cannot be added to mempool
func TestProcessedRequest(t *testing.T) {
	log := testlogger.NewLogger(t)
	glb := coreutil.NewGlobalSync().SetSolidIndex(0)
	rdr, vs := createStateReader(t, glb)
	wrt := vs.KVStore()

	pool := New(rdr, coretypes.NewInMemoryBlobCache(), log)
	require.NotNil(t, pool)

	stats := pool.Stats()
	require.EqualValues(t, 0, stats.InPoolCounter)
	require.EqualValues(t, 0, stats.OutPoolCounter)
	require.EqualValues(t, 0, stats.TotalPool)
	require.EqualValues(t, 0, stats.Ready)

	requests, _ := getRequestsOnLedger(t, 1)

	// artificially put request log record into the state
	rec := &blocklog.RequestLogRecord{
		RequestID: requests[0].ID(),
	}
	blocklogPartition := subrealm.New(wrt, kv.Key(blocklog.Interface.Hname().Bytes()))
	err := blocklog.SaveRequestLogRecord(blocklogPartition, rec, [6]byte{})
	require.NoError(t, err)
	blocklogPartition.Set(coreutil.StateVarBlockIndex, util.Uint64To8Bytes(1))
	err = vs.Commit()
	require.NoError(t, err)

	pool.ReceiveRequests(requests[0])
	require.False(t, pool.WaitRequestInPool(requests[0].ID(), 1*time.Second))

	stats = pool.Stats()
	require.EqualValues(t, 0, stats.InPoolCounter)
	require.EqualValues(t, 0, stats.OutPoolCounter)
	require.EqualValues(t, 0, stats.TotalPool)
	require.EqualValues(t, 0, stats.Ready)
}

//Test if adding and removing requests is handled correctly
func TestAddRemoveRequests(t *testing.T) {
	log := testlogger.NewLogger(t)
	glb := coreutil.NewGlobalSync().SetSolidIndex(0)
	rdr, _ := createStateReader(t, glb)
	pool := New(rdr, coretypes.NewInMemoryBlobCache(), log)
	require.NotNil(t, pool)
	requests, _ := getRequestsOnLedger(t, 6)

	pool.ReceiveRequests(
		requests[0],
		requests[1],
		requests[2],
		requests[3],
		requests[4],
		requests[5],
	)
	require.True(t, pool.WaitRequestInPool(requests[0].ID()))
	require.True(t, pool.WaitRequestInPool(requests[1].ID()))
	require.True(t, pool.WaitRequestInPool(requests[2].ID()))
	require.True(t, pool.WaitRequestInPool(requests[3].ID()))
	require.True(t, pool.WaitRequestInPool(requests[4].ID()))
	require.True(t, pool.WaitRequestInPool(requests[5].ID()))
	stats := pool.Stats()
	require.EqualValues(t, 6, stats.InPoolCounter)
	require.EqualValues(t, 0, stats.OutPoolCounter)
	require.EqualValues(t, 6, stats.TotalPool)
	require.EqualValues(t, 6, stats.Ready)

	pool.RemoveRequests(
		requests[3].ID(),
		requests[0].ID(),
		requests[1].ID(),
		requests[5].ID(),
	)
	require.False(t, pool.HasRequest(requests[0].ID()))
	require.False(t, pool.HasRequest(requests[1].ID()))
	require.True(t, pool.HasRequest(requests[2].ID()))
	require.False(t, pool.HasRequest(requests[3].ID()))
	require.True(t, pool.HasRequest(requests[4].ID()))
	require.False(t, pool.HasRequest(requests[5].ID()))
	stats = pool.Stats()
	require.EqualValues(t, 6, stats.InPoolCounter)
	require.EqualValues(t, 4, stats.OutPoolCounter)
	require.EqualValues(t, 2, stats.TotalPool)
	require.EqualValues(t, 2, stats.Ready)
}

//Test if ReadyNow and ReadyFromIDs functions respect the time lock of the request
func TestTimeLock(t *testing.T) {
	glb := coreutil.NewGlobalSync().SetSolidIndex(0)
	rdr, _ := createStateReader(t, glb)
	pool := New(rdr, coretypes.NewInMemoryBlobCache(), testlogger.NewLogger(t))
	require.NotNil(t, pool)
	requests, _ := getRequestsOnLedger(t, 6)

	now := time.Now()
	requests[1].Output().(*ledgerstate.ExtendedLockedOutput).WithTimeLock(now.Add(-2 * time.Hour))
	requests[2].Output().(*ledgerstate.ExtendedLockedOutput).WithTimeLock(now)
	requests[3].Output().(*ledgerstate.ExtendedLockedOutput).WithTimeLock(now.Add(2 * time.Hour))

	testStatsFun := func() { // Stats does not change after requests are added to the mempool
		stats := pool.Stats()
		require.EqualValues(t, 4, stats.InPoolCounter)
		require.EqualValues(t, 0, stats.OutPoolCounter)
		require.EqualValues(t, 4, stats.TotalPool)
		require.EqualValues(t, 3, stats.Ready)
	}
	pool.ReceiveRequests(
		requests[0], // + No time lock
		requests[1], // + Time lock before now
		requests[2], // + Time lock slightly before now due to time.Now() in ReadyNow being called later than in this test
		requests[3], // - Time lock after now
	)
	require.True(t, pool.WaitRequestInPool(requests[0].ID()))
	require.True(t, pool.WaitRequestInPool(requests[1].ID()))
	require.True(t, pool.WaitRequestInPool(requests[2].ID()))
	require.True(t, pool.WaitRequestInPool(requests[3].ID()))
	testStatsFun()

	ready := pool.ReadyNow()
	require.True(t, len(ready) == 3)
	require.Contains(t, ready, requests[0])
	require.Contains(t, ready, requests[1])
	require.Contains(t, ready, requests[2])
	testStatsFun()

	ready, result := pool.ReadyFromIDs(now.Add(-3*time.Hour),
		requests[0].ID(), // + No time lock
		requests[1].ID(), // - Time lock less than three hours before now
		requests[2].ID(), // - Time lock at exactly the same time as now
		requests[3].ID(), // - Time lock after now
	)
	require.True(t, result)
	require.True(t, len(ready) == 1)
	require.Contains(t, ready, requests[0])
	testStatsFun()

	ready, result = pool.ReadyFromIDs(now.Add(-1*time.Hour),
		requests[0].ID(), // + No time lock
		requests[1].ID(), // + Time lock more than one hour before now
		requests[2].ID(), // - Time lock at exactly the same time as now
		requests[3].ID(), // - Time lock after now
	)
	require.True(t, result)
	require.True(t, len(ready) == 2)
	require.Contains(t, ready, requests[0])
	require.Contains(t, ready, requests[1])
	testStatsFun()

	ready, result = pool.ReadyFromIDs(now,
		requests[0].ID(), // + No time lock
		requests[1].ID(), // + Time lock before now
		requests[2].ID(), // - Time lock at exactly the same time as now
		requests[3].ID(), // - Time lock after now
	)
	require.True(t, result)
	require.True(t, len(ready) == 2)
	require.Contains(t, ready, requests[0])
	require.Contains(t, ready, requests[1])
	testStatsFun()

	ready, result = pool.ReadyFromIDs(now.Add(1*time.Hour),
		requests[0].ID(), // + No time lock
		requests[1].ID(), // + Time lock before now
		requests[2].ID(), // + Time lock at exactly the same time as now
		requests[3].ID(), // - Time lock more than one hour after now
	)
	require.True(t, result)
	require.True(t, len(ready) == 3)
	require.Contains(t, ready, requests[0])
	require.Contains(t, ready, requests[1])
	require.Contains(t, ready, requests[2])
	testStatsFun()

	ready, result = pool.ReadyFromIDs(now.Add(3*time.Hour),
		requests[0].ID(), // + No time lock
		requests[1].ID(), // + Time lock before now
		requests[2].ID(), // + Time lock at exactly the same time as now
		requests[3].ID(), // + Time lock less than three hours after now
	)
	require.True(t, result)
	require.True(t, len(ready) == 4)
	require.Contains(t, ready, requests[0])
	require.Contains(t, ready, requests[1])
	require.Contains(t, ready, requests[2])
	require.Contains(t, ready, requests[3])
	testStatsFun()
}

//Test if ReadyFromIDs function correctly handle non-existing or removed IDs
func TestReadyFromIDs(t *testing.T) {
	glb := coreutil.NewGlobalSync().SetSolidIndex(0)
	rdr, _ := createStateReader(t, glb)
	pool := New(rdr, coretypes.NewInMemoryBlobCache(), testlogger.NewLogger(t))
	require.NotNil(t, pool)
	requests, _ := getRequestsOnLedger(t, 6)

	pool.ReceiveRequests(
		requests[0],
		requests[1],
		requests[2],
		requests[3],
		requests[4],
	)
	require.True(t, pool.WaitRequestInPool(requests[0].ID()))
	require.True(t, pool.WaitRequestInPool(requests[1].ID()))
	require.True(t, pool.WaitRequestInPool(requests[2].ID()))
	require.True(t, pool.WaitRequestInPool(requests[3].ID()))
	require.True(t, pool.WaitRequestInPool(requests[4].ID()))
	stats := pool.Stats()
	require.EqualValues(t, 5, stats.InPoolCounter)
	require.EqualValues(t, 0, stats.OutPoolCounter)
	require.EqualValues(t, 5, stats.TotalPool)
	require.EqualValues(t, 5, stats.Ready)

	ready, result := pool.ReadyFromIDs(time.Now(),
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
	stats = pool.Stats()
	require.EqualValues(t, 5, stats.InPoolCounter)
	require.EqualValues(t, 0, stats.OutPoolCounter)
	require.EqualValues(t, 5, stats.TotalPool)
	require.EqualValues(t, 5, stats.Ready)

	pool.RemoveRequests(requests[3].ID())
	_, result = pool.ReadyFromIDs(time.Now(),
		requests[0].ID(),
		requests[1].ID(),
		requests[2].ID(),
		requests[3].ID(), // Request was removed from mempool
	)
	require.False(t, result)
	_, result = pool.ReadyFromIDs(time.Now(),
		requests[5].ID(), // Request hasn't been received by mempool
		requests[4].ID(),
		requests[2].ID(),
	)
	require.False(t, result)
	ready, result = pool.ReadyFromIDs(time.Now(),
		requests[0].ID(),
		requests[1].ID(),
		requests[2].ID(),
		requests[4].ID(),
	)
	require.True(t, result)
	require.True(t, len(ready) == 4)
	require.Contains(t, ready, requests[0])
	require.Contains(t, ready, requests[1])
	require.Contains(t, ready, requests[2])
	require.Contains(t, ready, requests[4])
	stats = pool.Stats()
	require.EqualValues(t, 5, stats.InPoolCounter)
	require.EqualValues(t, 1, stats.OutPoolCounter)
	require.EqualValues(t, 4, stats.TotalPool)
	require.EqualValues(t, 4, stats.Ready)
}

//Test if solidification works as expected
func TestSolidification(t *testing.T) {
	log := testlogger.NewLogger(t)
	glb := coreutil.NewGlobalSync().SetSolidIndex(0)
	rdr, _ := createStateReader(t, glb)
	blobCache := coretypes.NewInMemoryBlobCache()
	pool := New(rdr, blobCache, log, 20*time.Millisecond) // Solidification initiated on pool creation
	require.NotNil(t, pool)
	requests, _ := getRequestsOnLedger(t, 4)

	// we need a request that actually requires solidification
	// because ReceiveRequest will already attempt to solidify
	blobData := []byte("blobData")
	args := requestargs.New(nil)
	hash := args.AddAsBlobRef("blob", blobData)
	meta := request.NewRequestMetadata().WithArgs(args)
	requests[0].SetMetadata(meta)

	// no solidification yet => request is not ready
	pool.ReceiveRequests(requests[0])
	require.True(t, pool.WaitRequestInPool(requests[0].ID()))
	ready, result := pool.ReadyFromIDs(time.Now(), requests[0].ID())
	require.True(t, result)
	require.True(t, len(ready) == 0)

	// provide the blob data in the blob cache
	blob, err := blobCache.PutBlob(blobData)
	require.NoError(t, err)
	require.EqualValues(t, blob, hash)

	// wait for solidification loop to initiate solidification
	time.Sleep(50 * time.Millisecond)

	// now that solidification happened => request is ready
	require.True(t, pool.HasRequest(requests[0].ID()))
	ready, result = pool.ReadyFromIDs(time.Now(), requests[0].ID())
	require.True(t, result)
	require.True(t, len(ready) == 1)
	require.Contains(t, ready, requests[0])
}

func TestRotateRequest(t *testing.T) {
	glb := coreutil.NewGlobalSync().SetSolidIndex(0)
	rdr, _ := createStateReader(t, glb)
	pool := New(rdr, coretypes.NewInMemoryBlobCache(), testlogger.NewLogger(t))
	require.NotNil(t, pool)
	requests, _ := getRequestsOnLedger(t, 6)

	pool.ReceiveRequests(
		requests[0],
		requests[1],
		requests[2],
		requests[3],
		requests[4],
	)
	require.True(t, pool.WaitRequestInPool(requests[0].ID()))
	require.True(t, pool.WaitRequestInPool(requests[1].ID()))
	require.True(t, pool.WaitRequestInPool(requests[2].ID()))
	require.True(t, pool.WaitRequestInPool(requests[3].ID()))
	require.True(t, pool.WaitRequestInPool(requests[4].ID()))
	stats := pool.Stats()
	require.EqualValues(t, 5, stats.InPoolCounter)
	require.EqualValues(t, 0, stats.OutPoolCounter)
	require.EqualValues(t, 5, stats.TotalPool)
	require.EqualValues(t, 5, stats.Ready)

	ready, result := pool.ReadyFromIDs(time.Now(),
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
	stats = pool.Stats()
	require.EqualValues(t, 5, stats.InPoolCounter)
	require.EqualValues(t, 0, stats.OutPoolCounter)
	require.EqualValues(t, 5, stats.TotalPool)
	require.EqualValues(t, 5, stats.Ready)

	pool.RemoveRequests(requests[3].ID())
	_, result = pool.ReadyFromIDs(time.Now(),
		requests[0].ID(),
		requests[1].ID(),
		requests[2].ID(),
		requests[3].ID(), // Request was removed from mempool
	)
	require.False(t, result)
	_, result = pool.ReadyFromIDs(time.Now(),
		requests[5].ID(), // Request hasn't been received by mempool
		requests[4].ID(),
		requests[2].ID(),
	)
	require.False(t, result)
	ready, result = pool.ReadyFromIDs(time.Now(),
		requests[0].ID(),
		requests[1].ID(),
		requests[2].ID(),
		requests[4].ID(),
	)
	require.True(t, result)
	require.True(t, len(ready) == 4)
	require.Contains(t, ready, requests[0])
	require.Contains(t, ready, requests[1])
	require.Contains(t, ready, requests[2])
	require.Contains(t, ready, requests[4])
	stats = pool.Stats()
	require.EqualValues(t, 5, stats.InPoolCounter)
	require.EqualValues(t, 1, stats.OutPoolCounter)
	require.EqualValues(t, 4, stats.TotalPool)
	require.EqualValues(t, 4, stats.Ready)
}
