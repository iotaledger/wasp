package mempool

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/hive.go/core/kvstore/mapdb"
	iotago "github.com/iotaledger/iota.go/v3"
	"github.com/iotaledger/iota.go/v3/tpkg"
	"github.com/iotaledger/wasp/packages/cryptolib"
	"github.com/iotaledger/wasp/packages/gpa"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/subrealm"
	"github.com/iotaledger/wasp/packages/state"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
	"github.com/iotaledger/wasp/packages/transaction"
	"github.com/iotaledger/wasp/packages/util"
	"github.com/iotaledger/wasp/packages/vm/core/blocklog"
)

var chainAddress = tpkg.RandAliasAddress()

func newTestPool(t *testing.T) Mempool {
	log := testlogger.NewLogger(t)
	var peeringNetwork *testutil.PeeringNetwork = testutil.NewPeeringNetwork(
		[]string{"nodeID"}, []*cryptolib.KeyPair{cryptolib.NewKeyPair()}, 10000,
		testutil.NewPeeringNetReliable(log),
		log,
	)

	glb := coreutil.NewChainStateSync().SetSolidIndex(0)
	stateReader, _ := createStateReader(t, glb)
	mempoolMetrics := new(MockMempoolMetrics)
	chainID := isc.ChainIDFromAddress(chainAddress)
	return New(
		context.Background(),
		&chainID,
		gpa.NodeID("nodeID"),
		peeringNetwork.NetworkProviders()[0],
		CreateHasBeenProcessedFunc(stateReader.KVStoreReader()),
		CreateGetProcessedReqsFunc(stateReader.KVStoreReader()),
		log,
		mempoolMetrics,
	)
}

func createStateReader(t *testing.T, glb coreutil.ChainStateSync) (state.OptimisticStateReader, state.VirtualStateAccess) {
	store := mapdb.NewMapDB()
	vs, err := state.CreateOriginState(store, isc.RandomChainID())
	require.NoError(t, err)
	ret := state.NewOptimisticStateReader(store, glb)
	require.NoError(t, err)
	return ret, vs
}

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

func (m *MockMempoolMetrics) CountRequestIn(req isc.Request) {
	if req.IsOffLedger() {
		m.offLedgerRequestCounter++
	} else {
		m.onLedgerRequestCounter++
	}
}

func (m *MockMempoolMetrics) CountRequestOut() {
	m.processedRequestCounter++
}

func (m *MockMempoolMetrics) RecordRequestProcessingTime(reqID isc.RequestID, elapse time.Duration) {
}

func (m *MockMempoolMetrics) CountBlocksPerChain() {}

// Test if mempool is created
func TestMempool(t *testing.T) {
	pool := newTestPool(t)
	stats := pool.Info()
	require.EqualValues(t, 0, stats.InPoolCounter)
	require.EqualValues(t, 0, stats.OutPoolCounter)
	require.EqualValues(t, 0, stats.TotalPool)
}

// Test if single on ledger request is added to mempool
func TestAddRequest(t *testing.T) {
	pool := newTestPool(t)
	requests := getRequestsOnLedger(t, 1)

	pool.ReceiveRequests(requests[0])
	time.Sleep(10 * time.Millisecond)
	// require.True(t, pool.WaitRequestInPool(requests[0].ID()))
	stats := pool.Info()
	require.EqualValues(t, 1, stats.InPoolCounter)
	require.EqualValues(t, 0, stats.OutPoolCounter)
	require.EqualValues(t, 1, stats.TotalPool)
	metrics := pool.(*mempool).metrics.(*MockMempoolMetrics)
	require.EqualValues(t, 1, metrics.onLedgerRequestCounter)
}

// Test if adding the same on ledger request more than once to the same mempool
// is handled correctly
func TestAddRequestTwice(t *testing.T) {
	pool := newTestPool(t)
	requests := getRequestsOnLedger(t, 1)

	pool.ReceiveRequests(requests[0])
	// require.True(t, pool.WaitRequestInPool(requests[0].ID(), 200*time.Millisecond))

	stats := pool.Info()
	require.EqualValues(t, 1, stats.InPoolCounter)
	require.EqualValues(t, 0, stats.OutPoolCounter)
	require.EqualValues(t, 1, stats.TotalPool)

	pool.ReceiveRequests(requests[0])
	// require.True(t, pool.WaitRequestInPool(requests[0].ID(), 200*time.Millisecond))

	stats = pool.Info()
	require.EqualValues(t, 1, stats.InPoolCounter)
	require.EqualValues(t, 0, stats.OutPoolCounter)
	require.EqualValues(t, 1, stats.TotalPool)
}

// Test if adding off ledger requests works as expected
func TestAddOffLedgerRequest(t *testing.T) {
	pool := newTestPool(t)
	offLedgerRequest := isc.NewOffLedgerRequest(isc.RandomChainID(), isc.Hn("dummyContract"), isc.Hn("dummyEP"), dict.New(), 0).
		Sign(cryptolib.NewKeyPair())
	metrics := pool.(*mempool).metrics.(*MockMempoolMetrics)
	require.EqualValues(t, 0, metrics.offLedgerRequestCounter)
	pool.ReceiveRequests(offLedgerRequest)
	// require.True(t, pool.WaitRequestInPool(offLedgerRequest.ID(), 200*time.Millisecond))
	stats := pool.Info()
	require.EqualValues(t, 1, stats.InPoolCounter)
	require.EqualValues(t, 0, stats.OutPoolCounter)
	require.EqualValues(t, 1, stats.TotalPool)
	require.EqualValues(t, 1, metrics.offLedgerRequestCounter)
}

// Test if processed request cannot be added to mempool
func TestProcessedRequest(t *testing.T) {
	pool := newTestPool(t)
	glb := coreutil.NewChainStateSync().SetSolidIndex(0)
	stateReader, vs := createStateReader(t, glb)
	pool.(*mempool).getProcessedRequests = CreateGetProcessedReqsFunc(stateReader.KVStoreReader())
	pool.(*mempool).hasBeenProcessed = CreateHasBeenProcessedFunc(stateReader.KVStoreReader())

	wrt := vs.KVStore()
	stats := pool.Info()
	require.EqualValues(t, 0, stats.InPoolCounter)
	require.EqualValues(t, 0, stats.OutPoolCounter)
	require.EqualValues(t, 0, stats.TotalPool)

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

	ret := pool.ReceiveRequests(requests[0])
	require.Len(t, ret, 1)
	require.False(t, ret[0])

	stats = pool.Info()
	require.EqualValues(t, 0, stats.InPoolCounter)
	require.EqualValues(t, 0, stats.OutPoolCounter)
	require.EqualValues(t, 0, stats.TotalPool)
}

// Test if adding and removing requests is handled correctly
func TestAddRemoveRequests(t *testing.T) {
	pool := newTestPool(t)
	requests := getRequestsOnLedger(t, 6)

	pool.ReceiveRequests(
		requests[0],
		requests[1],
		requests[2],
		requests[3],
		requests[4],
		requests[5],
	)
	stats := pool.Info()
	require.EqualValues(t, 6, stats.InPoolCounter)
	require.EqualValues(t, 0, stats.OutPoolCounter)
	require.EqualValues(t, 6, stats.TotalPool)

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
	metrics := pool.(*mempool).metrics.(*MockMempoolMetrics)
	require.EqualValues(t, 4, metrics.processedRequestCounter)
}

var mockAliasOutput = isc.NewAliasOutputWithID(&iotago.AliasOutput{}, nil)

func TestTimeLock(t *testing.T) {
	pool := newTestPool(t)
	start := time.Now()
	requests := getRequestsOnLedger(t, 6, func(i int, p *isc.RequestParameters) {
		switch i {
		case 1:
			p.Options.Timelock = start.Add(-2 * time.Hour)
		case 2:
			p.Options.Timelock = start
		case 3:
			p.Options.Timelock = start.Add(5 * time.Second)
		case 4:
			p.Options.Timelock = start.Add(2 * time.Hour)
		case 5:
			// expires before timelock
			p.Options.Timelock = start.Add(3 * time.Second)
			p.Options.Expiration = &isc.Expiration{
				Time:          start.Add(2 * time.Second),
				ReturnAddress: chainAddress,
			}
		}
	})

	testStatsFun := func(in, out, total int) { // Info does not change after requests are added to the mempool
		stats := pool.Info()
		require.EqualValues(t, in, stats.InPoolCounter)
		require.EqualValues(t, out, stats.OutPoolCounter)
		require.EqualValues(t, total, stats.TotalPool)
	}
	ret := pool.ReceiveRequests(
		requests[0], // + No time lock
		requests[1], // + Time lock before start
		requests[2], // + Time lock slightly before start due to time.Now() in ReadyNow being called later than in this test
		requests[3], // - Time lock 5s after start
		requests[4], // - Time lock 2h after start
		requests[5], // - Time lock after expiration
	)
	require.Len(t, ret, 6)
	require.True(t, ret[0])
	require.True(t, ret[1])
	require.True(t, ret[2])
	require.True(t, ret[3])
	require.True(t, ret[4])
	require.False(t, ret[5])
	testStatsFun(3, 0, 3)

	requestRefs := <-pool.ConsensusProposalsAsync(context.Background(), mockAliasOutput)
	requestsReady := <-pool.ConsensusRequestsAsync(context.Background(), requestRefs)

	require.Len(t, requestsReady, 3)
	require.Contains(t, requestsReady, requests[0])
	require.Contains(t, requestsReady, requests[1])
	require.Contains(t, requestsReady, requests[2])
	testStatsFun(3, 0, 3)

	// pass some time so that 1 request is unlocked
	time.Sleep(6 * time.Second)

	requestRefs = <-pool.ConsensusProposalsAsync(context.Background(), mockAliasOutput)
	requestsReady = <-pool.ConsensusRequestsAsync(context.Background(), requestRefs)

	require.Len(t, requestsReady, 4)
	require.Contains(t, requestsReady, requests[0])
	require.Contains(t, requestsReady, requests[1])
	require.Contains(t, requestsReady, requests[2])
	require.Contains(t, requestsReady, requests[3])
	testStatsFun(4, 0, 4)
}

func TestExpiration(t *testing.T) {
	pool := newTestPool(t)
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

	ret := pool.ReceiveRequests(
		requests[0], // + No expiration
		requests[1], // + Expired
		requests[2], // + Will expire soon
		requests[3], // + Still valid
	)
	require.True(t, ret[0])
	require.False(t, ret[1])
	require.False(t, ret[2])
	require.True(t, ret[3])

	stats := pool.Info()
	require.EqualValues(t, 2, stats.InPoolCounter)
	require.EqualValues(t, 0, stats.OutPoolCounter)
	require.EqualValues(t, 2, stats.TotalPool)

	requestRefs := <-pool.ConsensusProposalsAsync(context.Background(), mockAliasOutput)
	requestsReady := <-pool.ConsensusRequestsAsync(context.Background(), requestRefs)

	require.Len(t, requestsReady, 2)
	require.Contains(t, requestsReady, requests[0])
	require.Contains(t, requestsReady, requests[3])

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

// Test if ConsensusRequestsAsync function correctly handle non-existing or removed IDs
func TestConsensusRequestsAsync(t *testing.T) {
	pool := newTestPool(t)
	requests := getRequestsOnLedger(t, 5)

	res := pool.ReceiveRequests(
		requests[0],
		requests[1],
		requests[2],
		requests[3],
		requests[4],
	)
	require.Len(t, res, 5)
	require.True(t, res[0])
	require.True(t, res[1])
	require.True(t, res[2])
	require.True(t, res[3])
	require.True(t, res[4])

	stats := pool.Info()
	require.EqualValues(t, 5, stats.InPoolCounter)
	require.EqualValues(t, 0, stats.OutPoolCounter)
	require.EqualValues(t, 5, stats.TotalPool)

	requestRefs := <-pool.ConsensusProposalsAsync(context.Background(), mockAliasOutput)
	requestsReady := <-pool.ConsensusRequestsAsync(context.Background(), requestRefs)

	require.Len(t, requestsReady, 5)
	for _, req := range requests {
		require.Contains(t, requestsReady, req)
	}

	retReqs, missing := pool.(*mempool).getRequestsFromRefs(requestRefs)
	require.Empty(t, missing)
	require.Len(t, retReqs, 5)
	for _, req := range requests {
		require.Contains(t, retReqs, req)
	}

	// remove a request from the mempool
	pool.RemoveRequests(requests[3].ID())
	retReqs, missing = pool.(*mempool).getRequestsFromRefs(requestRefs)
	require.Len(t, missing, 1)
	require.NotNil(t, missing[isc.RequestHash(requests[3])])
	require.Len(t, retReqs, 5)
	for i, req := range requests {
		if i == 3 {
			require.NotContains(t, retReqs, req)
			continue
		}
		require.Contains(t, retReqs, req)
	}
}

func TestRequestsAreRemovedWithNewAliasOutput(t *testing.T) {
	pool := newTestPool(t)
	requests := getRequestsOnLedger(t, 2)

	pool.ReceiveRequests(
		requests[0],
		requests[1],
	)

	requestRefs := <-pool.ConsensusProposalsAsync(context.Background(), mockAliasOutput)
	requestsReady := <-pool.ConsensusRequestsAsync(context.Background(), requestRefs)

	require.Len(t, requestsReady, 2)
	for _, req := range requests {
		require.Contains(t, requestsReady, req)
	}

	// mock "state read" functions, so that request 0 has been proceeds, 1 not
	pool.(*mempool).getProcessedRequests = func(from, to *isc.AliasOutputWithID) []isc.RequestID {
		return []isc.RequestID{requests[0].ID()}
	}
	pool.(*mempool).hasBeenProcessed = func(reqID isc.RequestID) bool {
		return reqID.Equals(requests[0].ID())
	}

	nextAliasOutput := isc.NewAliasOutputWithID(&iotago.AliasOutput{
		StateIndex: mockAliasOutput.GetStateIndex() + 1,
	}, nil)
	requestRefs = <-pool.ConsensusProposalsAsync(context.Background(), nextAliasOutput)
	requestsReady = <-pool.ConsensusRequestsAsync(context.Background(), requestRefs)

	require.Len(t, requestsReady, 1)
	require.Equal(t, requestsReady[0], requests[1])
	// request0 has been removed from the pool
	require.False(t, pool.HasRequest(requests[0].ID()))
}
