package mempool

import (
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/core/kvstore/mapdb"
	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/isc/rotate"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/subrealm"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/testutil/testkey"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
)

var chainAddress = tpkg.RandEd25519Address()

func createStateReader(t *testing.T, glb coreutil.ChainStateSync) (state.OptimisticStateReader, state.VirtualStateAccess) {
	store := mapdb.NewMapDB()
	vs, err := state.CreateOriginState(store, isc.RandomChainID())
	require.NoError(t, err)
	ret := state.NewOptimisticStateReader(store, glb)
	require.NoError(t, err)
	return ret, vs
}

func now() time.Time { return time.Now() }

func getRequestsOnLedger(t *testing.T, amount int, f ...func(int, *isc.RequestParameters)) []isc.OnLedgerRequest {
	result := make([]isc.OnLedgerRequest, amount)
	for i := range result {
		requestParams := isc.RequestParameters{
			TargetAddress:  chainAddress,
			FungibleTokens: nil,
			Metadata: &isc.SendMetadata{
				TargetContract: isc.Hn("dummyTargetContract"),
				EntryPoint:     isc.Hn("dummyEP"),
				Params:         dict.New(),
				Allowance:      nil,
				GasBudget:      1000,
			},
			AdjustToMinimumStorageDeposit: true,
		}
		if len(f) == 1 {
			f[0](i, &requestParams)
		}
		output := transaction.BasicOutputFromPostData(
			tpkg.RandEd25519Address(),
			isc.Hn("dummySenderContract"),
			requestParams,
		)
		outputID := tpkg.RandOutputID(uint16(i)).UTXOInput()
		var err error
		result[i], err = isc.OnLedgerFromUTXO(output, outputID)
		require.NoError(t, err)
	}
	return result
}

type MockMempoolMetrics struct {
	mock.Mock
	offLedgerRequestCounter int
	onLedgerRequestCounter  int
	processedRequestCounter int
}

func (m *MockMempoolMetrics) CountOffLedgerRequestIn() {
	m.offLedgerRequestCounter++
}

func (m *MockMempoolMetrics) CountOnLedgerRequestIn() {
	m.onLedgerRequestCounter++
}

func (m *MockMempoolMetrics) CountRequestOut() {
	m.processedRequestCounter++
}

func (m *MockMempoolMetrics) RecordRequestProcessingTime(reqID isc.RequestID, elapse time.Duration) {
}

func (m *MockMempoolMetrics) CountBlocksPerChain() {}

// Test if mempool is created
func TestMempool(t *testing.T) {
	log := testlogger.NewLogger(t)
	glb := coreutil.NewChainStateSync()
	rdr, _ := createStateReader(t, glb)
	mempoolMetrics := new(MockMempoolMetrics)
	pool := New(chainAddress, rdr, log, mempoolMetrics)
	time.Sleep(2 * moveToPoolLoopDelay)
	stats := pool.Info(now())
	require.EqualValues(t, 0, stats.InPoolCounter)
	require.EqualValues(t, 0, stats.OutPoolCounter)
	require.EqualValues(t, 0, stats.TotalPool)
	require.EqualValues(t, 0, stats.ReadyCounter)
	pool.Close()
}

// Test if single on ledger request is added to mempool
func TestAddRequest(t *testing.T) {
	log := testlogger.NewLogger(t)
	glb := coreutil.NewChainStateSync().SetSolidIndex(0)
	rdr, _ := createStateReader(t, glb)
	mempoolMetrics := new(MockMempoolMetrics)
	pool := New(chainAddress, rdr, log, mempoolMetrics)
	requests := getRequestsOnLedger(t, 1)

	pool.ReceiveRequests(requests[0])
	require.True(t, pool.WaitRequestInPool(requests[0].ID()))
	stats := pool.Info(now())
	require.EqualValues(t, 1, stats.InPoolCounter)
	require.EqualValues(t, 0, stats.OutPoolCounter)
	require.EqualValues(t, 1, stats.TotalPool)
	require.EqualValues(t, 1, stats.ReadyCounter)
	require.EqualValues(t, 1, mempoolMetrics.onLedgerRequestCounter)
}

func TestAddRequestInvalidState(t *testing.T) {
	log := testlogger.NewLogger(t)
	glb := coreutil.NewChainStateSync()
	glb.InvalidateSolidIndex()
	rdr, _ := createStateReader(t, glb)
	mempoolMetrics := new(MockMempoolMetrics)
	pool := New(chainAddress, rdr, log, mempoolMetrics)
	requests := getRequestsOnLedger(t, 1)

	pool.ReceiveRequests(requests[0])
	require.False(t, pool.WaitRequestInPool(requests[0].ID(), 100*time.Millisecond))
	stats := pool.Info(now())
	require.EqualValues(t, 0, stats.InPoolCounter)
	require.EqualValues(t, 0, stats.OutPoolCounter)
	require.EqualValues(t, 0, stats.TotalPool)
	require.EqualValues(t, 0, stats.ReadyCounter)

	glb.SetSolidIndex(1)
	require.True(t, pool.WaitRequestInPool(requests[0].ID(), 100*time.Millisecond))
	stats = pool.Info(now())
	require.EqualValues(t, 1, stats.InPoolCounter)
	require.EqualValues(t, 0, stats.OutPoolCounter)
	require.EqualValues(t, 1, stats.TotalPool)
	require.EqualValues(t, 1, stats.ReadyCounter)
}

// Test if adding the same on ledger request more than once to the same mempool
// is handled correctly
func TestAddRequestTwice(t *testing.T) {
	log := testlogger.NewLogger(t)
	glb := coreutil.NewChainStateSync().SetSolidIndex(0)
	rdr, _ := createStateReader(t, glb)

	mempoolMetrics := new(MockMempoolMetrics)
	pool := New(chainAddress, rdr, log, mempoolMetrics)
	requests := getRequestsOnLedger(t, 1)

	pool.ReceiveRequests(requests[0])
	require.True(t, pool.WaitRequestInPool(requests[0].ID(), 200*time.Millisecond))

	stats := pool.Info(now())
	require.EqualValues(t, 1, stats.InPoolCounter)
	require.EqualValues(t, 0, stats.OutPoolCounter)
	require.EqualValues(t, 1, stats.TotalPool)
	require.EqualValues(t, 1, stats.ReadyCounter)

	pool.ReceiveRequests(requests[0])
	require.True(t, pool.WaitRequestInPool(requests[0].ID(), 200*time.Millisecond))

	stats = pool.Info(now())
	require.EqualValues(t, 1, stats.InPoolCounter)
	require.EqualValues(t, 0, stats.OutPoolCounter)
	require.EqualValues(t, 1, stats.TotalPool)
	require.EqualValues(t, 1, stats.ReadyCounter)
}

// Test if adding off ledger requests works as expected
func TestAddOffLedgerRequest(t *testing.T) {
	log := testlogger.NewLogger(t)
	glb := coreutil.NewChainStateSync().SetSolidIndex(0)
	rdr, _ := createStateReader(t, glb)
	mempoolMetrics := new(MockMempoolMetrics)
	pool := New(chainAddress, rdr, log, mempoolMetrics)

	offLedgerRequest := isc.NewOffLedgerRequest(isc.RandomChainID(), isc.Hn("dummyContract"), isc.Hn("dummyEP"), dict.New(), 0).
		Sign(cryptolib.NewKeyPair())
	require.EqualValues(t, 0, mempoolMetrics.offLedgerRequestCounter)
	pool.ReceiveRequests(offLedgerRequest)
	require.True(t, pool.WaitRequestInPool(offLedgerRequest.ID(), 200*time.Millisecond))
	stats := pool.Info(now())
	require.EqualValues(t, 1, stats.InPoolCounter)
	require.EqualValues(t, 0, stats.OutPoolCounter)
	require.EqualValues(t, 1, stats.TotalPool)
	require.EqualValues(t, 1, stats.ReadyCounter)
	require.EqualValues(t, 1, mempoolMetrics.offLedgerRequestCounter)
}

// Test if processed request cannot be added to mempool
func TestProcessedRequest(t *testing.T) {
	log := testlogger.NewLogger(t)
	glb := coreutil.NewChainStateSync().SetSolidIndex(0)
	rdr, vs := createStateReader(t, glb)
	wrt := vs.KVStore()

	mempoolMetrics := new(MockMempoolMetrics)
	pool := New(chainAddress, rdr, log, mempoolMetrics)

	stats := pool.Info(now())
	require.EqualValues(t, 0, stats.InPoolCounter)
	require.EqualValues(t, 0, stats.OutPoolCounter)
	require.EqualValues(t, 0, stats.TotalPool)
	require.EqualValues(t, 0, stats.ReadyCounter)

	requests := getRequestsOnLedger(t, 1)

	// artificially put request log record into the state
	rec := &blocklog.RequestReceipt{
		Request: requests[0],
	}
	blocklogPartition := subrealm.New(wrt, kv.Key(blocklog.Contract.Hname().Bytes()))
	err := blocklog.SaveRequestReceipt(blocklogPartition, rec, [6]byte{})
	require.NoError(t, err)
	blocklogPartition.Set(coreutil.StateVarBlockIndex, util.Uint64To8Bytes(1))
	err = vs.Save()
	require.NoError(t, err)

	pool.ReceiveRequests(requests[0])
	require.False(t, pool.WaitRequestInPool(requests[0].ID(), 1*time.Second))

	stats = pool.Info(now())
	require.EqualValues(t, 0, stats.InPoolCounter)
	require.EqualValues(t, 0, stats.OutPoolCounter)
	require.EqualValues(t, 0, stats.TotalPool)
	require.EqualValues(t, 0, stats.ReadyCounter)
}

// Test if adding and removing requests is handled correctly
func TestAddRemoveRequests(t *testing.T) {
	log := testlogger.NewLogger(t)
	glb := coreutil.NewChainStateSync().SetSolidIndex(0)
	rdr, _ := createStateReader(t, glb)
	mempoolMetrics := new(MockMempoolMetrics)
	pool := New(chainAddress, rdr, log, mempoolMetrics)
	requests := getRequestsOnLedger(t, 6)

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
	stats := pool.Info(now())
	require.EqualValues(t, 6, stats.InPoolCounter)
	require.EqualValues(t, 0, stats.OutPoolCounter)
	require.EqualValues(t, 6, stats.TotalPool)
	require.EqualValues(t, 6, stats.ReadyCounter)

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
	stats = pool.Info(now())
	require.EqualValues(t, 6, stats.InPoolCounter)
	require.EqualValues(t, 4, stats.OutPoolCounter)
	require.EqualValues(t, 2, stats.TotalPool)
	require.EqualValues(t, 2, stats.ReadyCounter)
	require.EqualValues(t, 4, mempoolMetrics.processedRequestCounter)
}

// Test if ReadyNow and ReadyFromIDs functions respect the time lock of the request
func TestTimeLock(t *testing.T) {
	glb := coreutil.NewChainStateSync().SetSolidIndex(0)
	rdr, _ := createStateReader(t, glb)
	mempoolMetrics := new(MockMempoolMetrics)
	pool := New(chainAddress, rdr, testlogger.NewLogger(t), mempoolMetrics)
	start := time.Now()
	requests := getRequestsOnLedger(t, 6, func(i int, p *isc.RequestParameters) {
		switch i {
		case 1:
			p.Options.Timelock = start.Add(-2 * time.Hour)
		case 2:
			p.Options.Timelock = start
		case 3:
			p.Options.Timelock = start.Add(2 * time.Hour)
		}
	})

	testStatsFun := func() { // Info does not change after requests are added to the mempool
		stats := pool.Info(start)
		require.EqualValues(t, 4, stats.InPoolCounter)
		require.EqualValues(t, 0, stats.OutPoolCounter)
		require.EqualValues(t, 4, stats.TotalPool)
		require.EqualValues(t, 3, stats.ReadyCounter)
	}
	pool.ReceiveRequests(
		requests[0], // + No time lock
		requests[1], // + Time lock before start
		requests[2], // + Time lock slightly before start due to time.Now() in ReadyNow being called later than in this test
		requests[3], // - Time lock after start
	)
	require.True(t, pool.WaitRequestInPool(requests[0].ID()))
	require.True(t, pool.WaitRequestInPool(requests[1].ID()))
	require.True(t, pool.WaitRequestInPool(requests[2].ID()))
	require.True(t, pool.WaitRequestInPool(requests[3].ID()))
	testStatsFun()

	ready, _, result := pool.ReadyFromIDs(start.Add(-3*time.Hour),
		requests[0].ID(), // + No time lock
		requests[1].ID(), // - Time lock less than three hours before start
		requests[2].ID(), // - Time lock at exactly the same time as start
		requests[3].ID(), // - Time lock after start
	)
	require.True(t, result)
	require.Len(t, ready, 1)
	require.Contains(t, ready, requests[0])
	testStatsFun()

	ready, _, result = pool.ReadyFromIDs(start.Add(-1*time.Hour),
		requests[0].ID(), // + No time lock
		requests[1].ID(), // + Time lock more than one hour before start
		requests[2].ID(), // - Time lock at exactly the same time as start
		requests[3].ID(), // - Time lock after start
	)
	require.True(t, result)
	require.Len(t, ready, 2)
	require.Contains(t, ready, requests[0])
	require.Contains(t, ready, requests[1])
	testStatsFun()

	ready, _, result = pool.ReadyFromIDs(start,
		requests[0].ID(), // + No time lock
		requests[1].ID(), // + Time lock before start
		requests[2].ID(), // - Time lock at exactly the same time as start
		requests[3].ID(), // - Time lock after start
	)
	require.True(t, result)
	require.Len(t, ready, 3)
	require.Contains(t, ready, requests[0])
	require.Contains(t, ready, requests[1])
	require.Contains(t, ready, requests[2])
	testStatsFun()

	ready, _, result = pool.ReadyFromIDs(start.Add(1*time.Hour),
		requests[0].ID(), // + No time lock
		requests[1].ID(), // + Time lock before start
		requests[2].ID(), // + Time lock at exactly the same time as start
		requests[3].ID(), // - Time lock more than one hour after start
	)
	require.True(t, result)
	require.Len(t, ready, 3)
	require.Contains(t, ready, requests[0])
	require.Contains(t, ready, requests[1])
	require.Contains(t, ready, requests[2])
	testStatsFun()

	ready, _, result = pool.ReadyFromIDs(start.Add(3*time.Hour),
		requests[0].ID(), // + No time lock
		requests[1].ID(), // + Time lock before start
		requests[2].ID(), // + Time lock at exactly the same time as start
		requests[3].ID(), // + Time lock less than three hours after start
	)
	require.True(t, result)
	require.Len(t, ready, 4)
	require.Contains(t, ready, requests[0])
	require.Contains(t, ready, requests[1])
	require.Contains(t, ready, requests[2])
	require.Contains(t, ready, requests[3])
	testStatsFun()
}

func TestExpiration(t *testing.T) {
	glb := coreutil.NewChainStateSync().SetSolidIndex(0)
	rdr, _ := createStateReader(t, glb)
	mempoolMetrics := new(MockMempoolMetrics)
	pool := New(chainAddress, rdr, testlogger.NewLogger(t), mempoolMetrics)
	start := time.Now()
	requests := getRequestsOnLedger(t, 4, func(i int, p *isc.RequestParameters) {
		switch i {
		case 1:
			// expired
			p.Options.Expiration = &isc.Expiration{
				Time:          start.Add(-isc.RequestConsideredExpiredWindow),
				ReturnAddress: chainAddress,
			}
		case 2:
			// will expire soon
			p.Options.Expiration = &isc.Expiration{
				Time:          start.Add(isc.RequestConsideredExpiredWindow / 2),
				ReturnAddress: chainAddress,
			}
		case 3:
			// not expired yet
			p.Options.Expiration = &isc.Expiration{
				Time:          start.Add(isc.RequestConsideredExpiredWindow * 2),
				ReturnAddress: chainAddress,
			}
		}
	})

	pool.ReceiveRequests(
		requests[0], // + No expiration
		requests[1], // + Expired
		requests[2], // + Will expire soon
		requests[3], // + Still valid
	)

	require.True(t, pool.WaitRequestInPool(requests[0].ID()))
	require.True(t, pool.WaitRequestInPool(requests[1].ID()))
	require.True(t, pool.WaitRequestInPool(requests[2].ID()))
	require.True(t, pool.WaitRequestInPool(requests[3].ID()))

	stats := pool.Info(start)
	require.EqualValues(t, 4, stats.InPoolCounter)
	require.EqualValues(t, 0, stats.OutPoolCounter)
	require.EqualValues(t, 4, stats.TotalPool)
	require.EqualValues(t, 2, stats.ReadyCounter)

	ready := pool.ReadyNow(start)
	require.Len(t, ready, 2)
	require.Contains(t, ready, requests[0])
	require.Contains(t, ready, requests[3])

	// requests with the invalid deadline should have been removed from the mempool
	ok := false
	for i := 0; i < 100; i++ {
		// just to let the `RemoveRequests` go routine get the pool mutex before we look into it
		time.Sleep(10 * time.Millisecond)
		if pool.GetRequest(requests[1].ID()) != nil {
			continue
		}
		if pool.GetRequest(requests[2].ID()) != nil {
			continue
		}
		if len(pool.(*mempool).pool) != 2 {
			continue
		}
		ok = true
		break
	}
	require.True(t, ok)
}

// Test if ReadyFromIDs function correctly handle non-existing or removed IDs
func TestReadyFromIDs(t *testing.T) {
	glb := coreutil.NewChainStateSync().SetSolidIndex(0)
	rdr, _ := createStateReader(t, glb)
	mempoolMetrics := new(MockMempoolMetrics)
	pool := New(chainAddress, rdr, testlogger.NewLogger(t), mempoolMetrics)
	requests := getRequestsOnLedger(t, 6)

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
	stats := pool.Info(now())
	require.EqualValues(t, 5, stats.InPoolCounter)
	require.EqualValues(t, 0, stats.OutPoolCounter)
	require.EqualValues(t, 5, stats.TotalPool)
	require.EqualValues(t, 5, stats.ReadyCounter)

	ready, missingIndexes, result := pool.ReadyFromIDs(now(),
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
	require.Empty(t, missingIndexes)
	stats = pool.Info(now())
	require.EqualValues(t, 5, stats.InPoolCounter)
	require.EqualValues(t, 0, stats.OutPoolCounter)
	require.EqualValues(t, 5, stats.TotalPool)
	require.EqualValues(t, 5, stats.ReadyCounter)

	pool.RemoveRequests(requests[3].ID())
	_, missingIndexes, result = pool.ReadyFromIDs(now(),
		requests[0].ID(),
		requests[1].ID(),
		requests[2].ID(),
		requests[3].ID(), // Request was removed from mempool
	)
	require.False(t, result)
	require.EqualValues(t, missingIndexes, []int{3})
	_, missingIndexes, result = pool.ReadyFromIDs(now(),
		requests[5].ID(), // Request hasn't been received by mempool
		requests[4].ID(),
		requests[2].ID(),
	)
	require.False(t, result)
	require.EqualValues(t, missingIndexes, []int{0})
	ready, _, result = pool.ReadyFromIDs(now(),
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
	stats = pool.Info(now())
	require.EqualValues(t, 5, stats.InPoolCounter)
	require.EqualValues(t, 1, stats.OutPoolCounter)
	require.EqualValues(t, 4, stats.TotalPool)
	require.EqualValues(t, 4, stats.ReadyCounter)
}

func TestRotateRequest(t *testing.T) {
	glb := coreutil.NewChainStateSync().SetSolidIndex(0)
	rdr, _ := createStateReader(t, glb)
	mempoolMetrics := new(MockMempoolMetrics)
	pool := New(chainAddress, rdr, testlogger.NewLogger(t), mempoolMetrics)
	requests := getRequestsOnLedger(t, 6)

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
	stats := pool.Info(now())
	require.EqualValues(t, 5, stats.InPoolCounter)
	require.EqualValues(t, 0, stats.OutPoolCounter)
	require.EqualValues(t, 5, stats.TotalPool)
	require.EqualValues(t, 5, stats.ReadyCounter)

	ready, _, result := pool.ReadyFromIDs(now(),
		requests[0].ID(),
		requests[1].ID(),
		requests[2].ID(),
		requests[3].ID(),
		requests[4].ID(),
	)
	require.True(t, result)
	require.True(t, len(ready) == 5)

	kp, addr := testkey.GenKeyAddr()
	rotateReq := rotate.NewRotateRequestOffLedger(isc.RandomChainID(), addr, kp)
	require.True(t, rotate.IsRotateStateControllerRequest(rotateReq))

	pool.ReceiveRequest(rotateReq)
	require.True(t, pool.WaitRequestInPool(rotateReq.ID()))
	require.True(t, pool.HasRequest(rotateReq.ID()))

	stats = pool.Info(now())
	require.EqualValues(t, 6, stats.TotalPool)

	ready = pool.ReadyNow(now())
	require.EqualValues(t, 1, len(ready))
	require.EqualValues(t, rotateReq.ID(), ready[0].ID())

	ready, _, ok := pool.ReadyFromIDs(now(), rotateReq.ID())
	require.True(t, ok)
	require.EqualValues(t, 1, len(ready))
	require.EqualValues(t, rotateReq.ID(), ready[0].ID())

	pool.RemoveRequests(rotateReq.ID())
	require.False(t, pool.HasRequest(rotateReq.ID()))

	ready = pool.ReadyNow(now())
	require.EqualValues(t, 5, len(ready))

	ready, _, result = pool.ReadyFromIDs(now(),
		requests[0].ID(),
		requests[1].ID(),
		requests[2].ID(),
		requests[3].ID(),
		requests[4].ID(),
	)
	require.True(t, result)
	require.True(t, len(ready) == 5)
}
