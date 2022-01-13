package mempool

import (
	"testing"
	"time"

	"github.com/iotaledger/hive.go/kvstore/mapdb"
	//iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/iscp/rotate"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/subrealm"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/testutil/testkey"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/utxodb"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func createStateReader(t *testing.T, glb coreutil.ChainStateSync) (state.OptimisticStateReader, state.VirtualStateAccess) {
	store := mapdb.NewMapDB()
	vs, err := state.CreateOriginState(store, nil)
	require.NoError(t, err)
	ret := state.NewOptimisticStateReader(store, glb)
	require.NoError(t, err)
	return ret, vs
}

func getRequestsOnLedger(t *testing.T, amount int) []*iscp.OnLedgerRequestData {
	utxo := utxodb.New()
	var err error
	addr := tpkg.RandEd25519Address()
	requestParams := iscp.RequestParameters{
		TargetAddress: tpkg.RandEd25519Address(),
		Assets:        nil,
		Metadata: &iscp.SendMetadata{
			TargetContract: iscp.Hn("dummyTargetContract"),
			EntryPoint:     iscp.Hn("dummyEP"),
			Params:         dict.New(),
			Allowance:      nil,
			GasBudget:      1000,
		},
	}
	result := make([]*iscp.OnLedgerRequestData, amount)
	for i := range result {
		output := transaction.ExtendedOutputFromPostData(addr, iscp.Hn("dummySenderContract"), requestParams, utxo.RentStructure())
		outputID := tpkg.RandOutputID(uint16(i)).UTXOInput()
		result[i], err = iscp.OnLedgerFromUTXO(output, outputID)
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

func (m *MockMempoolMetrics) RecordRequestProcessingTime(reqID iscp.RequestID, elapse time.Duration) {
}

func (m *MockMempoolMetrics) CountBlocksPerChain() {}

// Test if mempool is created
func TestMempool(t *testing.T) {
	log := testlogger.NewLogger(t)
	glb := coreutil.NewChainStateSync()
	rdr, _ := createStateReader(t, glb)
	mempoolMetrics := new(MockMempoolMetrics)
	pool := New(rdr, log, mempoolMetrics)
	require.NotNil(t, pool)
	time.Sleep(2 * time.Second)
	stats := pool.Info()
	require.EqualValues(t, 0, stats.InPoolCounter)
	require.EqualValues(t, 0, stats.OutPoolCounter)
	require.EqualValues(t, 0, stats.TotalPool)
	require.EqualValues(t, 0, stats.ReadyCounter)
	pool.Close()
	time.Sleep(1 * time.Second)
}

// Test if single on ledger request is added to mempool
func TestAddRequest(t *testing.T) {
	log := testlogger.NewLogger(t)
	glb := coreutil.NewChainStateSync().SetSolidIndex(0)
	rdr, _ := createStateReader(t, glb)
	mempoolMetrics := new(MockMempoolMetrics)
	pool := New(rdr, log, mempoolMetrics)
	require.NotNil(t, pool)
	requests := getRequestsOnLedger(t, 1)

	pool.ReceiveRequests(requests[0])
	require.True(t, pool.WaitRequestInPool(requests[0].ID()))
	stats := pool.Info()
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
	pool := New(rdr, log, mempoolMetrics)
	require.NotNil(t, pool)
	requests := getRequestsOnLedger(t, 1)

	pool.ReceiveRequests(requests[0])
	require.False(t, pool.WaitRequestInPool(requests[0].ID(), 100*time.Millisecond))
	stats := pool.Info()
	require.EqualValues(t, 0, stats.InPoolCounter)
	require.EqualValues(t, 0, stats.OutPoolCounter)
	require.EqualValues(t, 0, stats.TotalPool)
	require.EqualValues(t, 0, stats.ReadyCounter)

	glb.SetSolidIndex(1)
	require.True(t, pool.WaitRequestInPool(requests[0].ID(), 100*time.Millisecond))
	stats = pool.Info()
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
	pool := New(rdr, log, mempoolMetrics)
	require.NotNil(t, pool)
	requests := getRequestsOnLedger(t, 1)

	pool.ReceiveRequests(requests[0])
	require.True(t, pool.WaitRequestInPool(requests[0].ID(), 200*time.Millisecond))

	stats := pool.Info()
	require.EqualValues(t, 1, stats.InPoolCounter)
	require.EqualValues(t, 0, stats.OutPoolCounter)
	require.EqualValues(t, 1, stats.TotalPool)
	require.EqualValues(t, 1, stats.ReadyCounter)

	pool.ReceiveRequests(requests[0])
	require.True(t, pool.WaitRequestInPool(requests[0].ID(), 200*time.Millisecond))

	stats = pool.Info()
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
	pool := New(rdr, log, mempoolMetrics)
	require.NotNil(t, pool)

	offLedgerRequest := iscp.NewOffLedgerRequest(iscp.RandomChainID(), iscp.Hn("dummyContract"), iscp.Hn("dummyEP"), dict.New(), 0)
	require.EqualValues(t, 0, mempoolMetrics.offLedgerRequestCounter)
	pool.ReceiveRequests(offLedgerRequest)
	require.True(t, pool.WaitRequestInPool(offLedgerRequest.ID(), 200*time.Millisecond))
	stats := pool.Info()
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
	pool := New(rdr, log, mempoolMetrics)
	require.NotNil(t, pool)

	stats := pool.Info()
	require.EqualValues(t, 0, stats.InPoolCounter)
	require.EqualValues(t, 0, stats.OutPoolCounter)
	require.EqualValues(t, 0, stats.TotalPool)
	require.EqualValues(t, 0, stats.ReadyCounter)

	requests := getRequestsOnLedger(t, 1)

	// artificially put request log record into the state
	rec := &blocklog.RequestReceipt{
		RequestData: requests[0],
	}
	blocklogPartition := subrealm.New(wrt, kv.Key(blocklog.Contract.Hname().Bytes()))
	err := blocklog.SaveRequestReceipt(blocklogPartition, rec, [6]byte{})
	require.NoError(t, err)
	blocklogPartition.Set(coreutil.StateVarBlockIndex, util.Uint64To8Bytes(1))
	err = vs.Commit()
	require.NoError(t, err)

	pool.ReceiveRequests(requests[0])
	require.False(t, pool.WaitRequestInPool(requests[0].ID(), 1*time.Second))

	stats = pool.Info()
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
	pool := New(rdr, log, mempoolMetrics)
	require.NotNil(t, pool)
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
	stats := pool.Info()
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
	stats = pool.Info()
	require.EqualValues(t, 6, stats.InPoolCounter)
	require.EqualValues(t, 4, stats.OutPoolCounter)
	require.EqualValues(t, 2, stats.TotalPool)
	require.EqualValues(t, 2, stats.ReadyCounter)
	require.EqualValues(t, 4, mempoolMetrics.processedRequestCounter)
}

// Test if ReadyNow and ReadyFromIDs functions respect the time lock of the request
/*func TestTimeLock(t *testing.T) {
	glb := coreutil.NewChainStateSync().SetSolidIndex(0)
	rdr, _ := createStateReader(t, glb)
	mempoolMetrics := new(MockMempoolMetrics)
	pool := New(rdr, testlogger.NewLogger(t), mempoolMetrics)
	require.NotNil(t, pool)
	requests := getRequestsOnLedger(t, 6)

	now := time.Now()
	requests[1].Output().(*iotago.ExtendedOutput).WithTimeLock(now.Add(-2 * time.Hour))
	requests[2].Output().(*iotago.ExtendedOutput).WithTimeLock(now)
	requests[3].Output().(*iotago.ExtendedOutput).WithTimeLock(now.Add(2 * time.Hour))

	testStatsFun := func() { // Info does not change after requests are added to the mempool
		stats := pool.Info()
		require.EqualValues(t, 4, stats.InPoolCounter)
		require.EqualValues(t, 0, stats.OutPoolCounter)
		require.EqualValues(t, 4, stats.TotalPool)
		require.EqualValues(t, 3, stats.ReadyCounter)
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

	ready, _, result := pool.ReadyFromIDs(now.Add(-3*time.Hour),
		requests[0].ID(), // + No time lock
		requests[1].ID(), // - Time lock less than three hours before now
		requests[2].ID(), // - Time lock at exactly the same time as now
		requests[3].ID(), // - Time lock after now
	)
	require.True(t, result)
	require.True(t, len(ready) == 1)
	require.Contains(t, ready, requests[0])
	testStatsFun()

	ready, _, result = pool.ReadyFromIDs(now.Add(-1*time.Hour),
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

	ready, _, result = pool.ReadyFromIDs(now,
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

	ready, _, result = pool.ReadyFromIDs(now.Add(1*time.Hour),
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

	ready, _, result = pool.ReadyFromIDs(now.Add(3*time.Hour),
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
}*/

/*func TestFallbackOptions(t *testing.T) {
	glb := coreutil.NewChainStateSync().SetSolidIndex(0)
	rdr, _ := createStateReader(t, glb)
	mempoolMetrics := new(MockMempoolMetrics)
	pool := New(rdr, testlogger.NewLogger(t), mempoolMetrics)
	require.NotNil(t, pool)
	requests := getRequestsOnLedger(t, 3)

	address := ledgerstate.NewAliasAddress([]byte{1, 2, 3})
	validDeadline := time.Now().Add(FallbackDeadlineMinAllowedInterval).Add(time.Second)
	pastDeadline := time.Now().Add(-time.Second)
	requests[1].Output().(*iotago.ExtendedOutput).WithFallbackOptions(address, validDeadline)
	requests[2].Output().(*iotago.ExtendedOutput).WithFallbackOptions(address, pastDeadline)

	testStatsFun := func() { // Info does not change after requests are added to the mempool
		stats := pool.Info()
		require.EqualValues(t, 3, stats.InPoolCounter)
		require.EqualValues(t, 0, stats.OutPoolCounter)
		require.EqualValues(t, 3, stats.TotalPool)
		require.EqualValues(t, 2, stats.ReadyCounter)
	}

	pool.ReceiveRequests(
		requests[0], // + No fallback options
		requests[1], // + Valid deadline
		requests[2], // + Expired deadline
	)

	require.True(t, pool.WaitRequestInPool(requests[0].ID()))
	require.True(t, pool.WaitRequestInPool(requests[1].ID()))
	require.True(t, pool.WaitRequestInPool(requests[2].ID()))
	testStatsFun()

	ready := pool.ReadyNow()
	require.True(t, len(ready) == 2)
	require.Contains(t, ready, requests[0])
	require.Contains(t, ready, requests[1])

	// request with the invalid deadline should have been removed from the mempool
	time.Sleep(500 * time.Millisecond) // just to let the `RemoveRequests` go routine get the pool mutex before we look into it
	require.Nil(t, pool.GetRequest(requests[2].ID()))
	require.Len(t, pool.(*mempool).pool, 2)
}*/

// Test if ReadyFromIDs function correctly handle non-existing or removed IDs
func TestReadyFromIDs(t *testing.T) {
	glb := coreutil.NewChainStateSync().SetSolidIndex(0)
	rdr, _ := createStateReader(t, glb)
	mempoolMetrics := new(MockMempoolMetrics)
	pool := New(rdr, testlogger.NewLogger(t), mempoolMetrics)
	require.NotNil(t, pool)
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
	stats := pool.Info()
	require.EqualValues(t, 5, stats.InPoolCounter)
	require.EqualValues(t, 0, stats.OutPoolCounter)
	require.EqualValues(t, 5, stats.TotalPool)
	require.EqualValues(t, 5, stats.ReadyCounter)

	ready, missingIndexes, result := pool.ReadyFromIDs(time.Now(),
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
	stats = pool.Info()
	require.EqualValues(t, 5, stats.InPoolCounter)
	require.EqualValues(t, 0, stats.OutPoolCounter)
	require.EqualValues(t, 5, stats.TotalPool)
	require.EqualValues(t, 5, stats.ReadyCounter)

	pool.RemoveRequests(requests[3].ID())
	_, missingIndexes, result = pool.ReadyFromIDs(time.Now(),
		requests[0].ID(),
		requests[1].ID(),
		requests[2].ID(),
		requests[3].ID(), // Request was removed from mempool
	)
	require.False(t, result)
	require.EqualValues(t, missingIndexes, []int{3})
	_, missingIndexes, result = pool.ReadyFromIDs(time.Now(),
		requests[5].ID(), // Request hasn't been received by mempool
		requests[4].ID(),
		requests[2].ID(),
	)
	require.False(t, result)
	require.EqualValues(t, missingIndexes, []int{0})
	ready, _, result = pool.ReadyFromIDs(time.Now(),
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
	stats = pool.Info()
	require.EqualValues(t, 5, stats.InPoolCounter)
	require.EqualValues(t, 1, stats.OutPoolCounter)
	require.EqualValues(t, 4, stats.TotalPool)
	require.EqualValues(t, 4, stats.ReadyCounter)
}

func TestRotateRequest(t *testing.T) {
	glb := coreutil.NewChainStateSync().SetSolidIndex(0)
	rdr, _ := createStateReader(t, glb)
	mempoolMetrics := new(MockMempoolMetrics)
	pool := New(rdr, testlogger.NewLogger(t), mempoolMetrics)
	require.NotNil(t, pool)
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
	stats := pool.Info()
	require.EqualValues(t, 5, stats.InPoolCounter)
	require.EqualValues(t, 0, stats.OutPoolCounter)
	require.EqualValues(t, 5, stats.TotalPool)
	require.EqualValues(t, 5, stats.ReadyCounter)

	ready, _, result := pool.ReadyFromIDs(time.Now(),
		requests[0].ID(),
		requests[1].ID(),
		requests[2].ID(),
		requests[3].ID(),
		requests[4].ID(),
	)
	require.True(t, result)
	require.True(t, len(ready) == 5)

	kp, addr := testkey.GenKeyAddr()
	rotateReq := rotate.NewRotateRequestOffLedger(iscp.RandomChainID(), addr, *kp)
	require.True(t, rotate.IsRotateStateControllerRequest(rotateReq))

	pool.ReceiveRequest(rotateReq)
	require.True(t, pool.WaitRequestInPool(rotateReq.ID()))
	require.True(t, pool.HasRequest(rotateReq.ID()))

	stats = pool.Info()
	require.EqualValues(t, 6, stats.TotalPool)

	ready = pool.ReadyNow(time.Now())
	require.EqualValues(t, 1, len(ready))
	require.EqualValues(t, rotateReq.ID(), ready[0].ID())

	ready, _, ok := pool.ReadyFromIDs(time.Now(), rotateReq.ID())
	require.True(t, ok)
	require.EqualValues(t, 1, len(ready))
	require.EqualValues(t, rotateReq.ID(), ready[0].ID())

	pool.RemoveRequests(rotateReq.ID())
	require.False(t, pool.HasRequest(rotateReq.ID()))

	ready = pool.ReadyNow(time.Now())
	require.EqualValues(t, 5, len(ready))

	ready, _, result = pool.ReadyFromIDs(time.Now(),
		requests[0].ID(),
		requests[1].ID(),
		requests[2].ID(),
		requests[3].ID(),
		requests[4].ID(),
	)
	require.True(t, result)
	require.True(t, len(ready) == 5)
}
