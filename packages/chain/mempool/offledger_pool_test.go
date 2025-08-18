package mempool

import (
	"math/big"
	"testing"
	"time"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/packages/chain/cmtlog"
	consGR "github.com/iotaledger/wasp/v2/packages/chain/cons/gr"
	"github.com/iotaledger/wasp/v2/packages/cryptolib"
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/isc/isctest"
	"github.com/iotaledger/wasp/v2/packages/testutil"
	"github.com/iotaledger/wasp/v2/packages/testutil/testkey"
	"github.com/iotaledger/wasp/v2/packages/testutil/testlogger"
)

func TestOffledgerMempoolAccountNonce(t *testing.T) {
	waitReq := NewWaitReq(waitRequestCleanupEvery)
	pool := NewOffledgerPool(100, 10, waitReq, func(int) {}, func(time.Duration) {}, testlogger.NewSilentLogger("", true))

	// generate a bunch of requests for the same account
	kp, addr := testkey.GenKeyAddr()
	agentID := isc.NewAddressAgentID(addr)

	req0 := testutil.DummyOffledgerRequestForAccount(isctest.RandomChainID(), 0, kp)
	req1 := testutil.DummyOffledgerRequestForAccount(isctest.RandomChainID(), 1, kp)
	req2 := testutil.DummyOffledgerRequestForAccount(isctest.RandomChainID(), 2, kp)
	req2new := testutil.DummyOffledgerRequestForAccount(isctest.RandomChainID(), 2, kp)
	pool.Add(req0)
	pool.Add(req1)
	pool.Add(req1) // try to add the same request many times, old will be replaced.
	pool.Add(req2)
	pool.Add(req1)
	require.EqualValues(t, 3, pool.refLUT.Size())
	require.EqualValues(t, 1, pool.reqsByAcountOrdered.Size())
	reqsInPoolForAccount, _ := pool.reqsByAcountOrdered.Get(agentID.String())
	require.Len(t, reqsInPoolForAccount, 3)
	// Mark existing requests as proposed.
	consLogIndex := cmtlog.NilLogIndex()
	consID := consGR.NewConsensusID(cryptolib.NewEmptyAddress(), &consLogIndex)
	lo.ForEach(pool.orderedByGasPrice, func(e *OrderedPoolEntry, _ int) { e.markProposed(consID) })
	// Add it again. It should not be replaced, but appended instead.
	pool.Add(req2new)
	require.EqualValues(t, 4, pool.refLUT.Size())
	require.EqualValues(t, 1, pool.reqsByAcountOrdered.Size())
	reqsInPoolForAccount, _ = pool.reqsByAcountOrdered.Get(agentID.String())
	require.Len(t, reqsInPoolForAccount, 4)

	// try to remove everything during iteration
	pool.Iterate(func(account string, entries []*OrderedPoolEntry) bool {
		for _, e := range entries {
			pool.Remove(e.req)
		}

		return true
	})
	require.EqualValues(t, 0, pool.refLUT.Size())
	require.EqualValues(t, 0, pool.reqsByAcountOrdered.Size())
}

func TestOffledgerMempoolLimit(t *testing.T) {
	waitReq := NewWaitReq(waitRequestCleanupEvery)
	poolSizeLimit := 3
	pool := NewOffledgerPool(poolSizeLimit, poolSizeLimit, waitReq, func(int) {}, func(time.Duration) {}, testlogger.NewSilentLogger("", true))

	// create requests with different gas prices
	req0 := testutil.DummyEVMRequest(isctest.RandomChainID(), big.NewInt(1))
	req1 := testutil.DummyEVMRequest(isctest.RandomChainID(), big.NewInt(2))
	req2 := testutil.DummyEVMRequest(isctest.RandomChainID(), big.NewInt(3))
	pool.Add(req0)
	pool.Add(req1)
	pool.Add(req2)

	assertPoolSize := func() {
		require.EqualValues(t, 3, pool.refLUT.Size())
		require.Len(t, pool.orderedByGasPrice, 3)
		require.EqualValues(t, 3, pool.reqsByAcountOrdered.Size())
	}
	contains := func(reqs ...isc.OffLedgerRequest) {
		for _, req := range reqs {
			lo.ContainsBy(pool.orderedByGasPrice, func(e *OrderedPoolEntry) bool {
				return e.req.ID().Equals(req.ID())
			})
		}
	}

	assertPoolSize()

	// add a request with high
	req3 := testutil.DummyEVMRequest(isctest.RandomChainID(), big.NewInt(3))
	pool.Add(req3)
	assertPoolSize()
	contains(req1, req2, req3) // assert req3 was added and req0 was removed

	req4 := testutil.DummyEVMRequest(isctest.RandomChainID(), big.NewInt(1))
	pool.Add(req4)
	assertPoolSize()
	contains(req1, req2, req3) // assert req4 is not added

	req5 := testutil.DummyEVMRequest(isctest.RandomChainID(), big.NewInt(4))
	pool.Add(req5)
	assertPoolSize()

	contains(req2, req3, req5) // assert req5 was added and req1 was removed
}
