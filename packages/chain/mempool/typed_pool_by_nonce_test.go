package mempool

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/testutil"
	"github.com/iotaledger/wasp/packages/testutil/testkey"
	"github.com/iotaledger/wasp/packages/testutil/testlogger"
)

func TestSomething(t *testing.T) {
	waitReq := NewWaitReq(waitRequestCleanupEvery)
	pool := NewTypedPoolByNonce[isc.OffLedgerRequest](waitReq, func(int) {}, func(time.Duration) {}, testlogger.NewSilentLogger("", true))

	// generate a bunch of requests for the same account
	kp, addr := testkey.GenKeyAddr()
	agentID := isc.NewAgentID(addr)

	req0 := testutil.DummyOffledgerRequestForAccount(isc.RandomChainID(), 0, kp)
	req1 := testutil.DummyOffledgerRequestForAccount(isc.RandomChainID(), 1, kp)
	req2 := testutil.DummyOffledgerRequestForAccount(isc.RandomChainID(), 2, kp)
	req2new := testutil.DummyOffledgerRequestForAccount(isc.RandomChainID(), 2, kp)
	pool.Add(req0)
	pool.Add(req1)
	pool.Add(req1) // try to add the same request many times
	pool.Add(req2)
	pool.Add(req1)
	require.EqualValues(t, 3, pool.refLUT.Size())
	require.EqualValues(t, 1, pool.reqsByAcountOrdered.Size())
	reqsInPoolForAccount, _ := pool.reqsByAcountOrdered.Get(agentID.String())
	require.Len(t, reqsInPoolForAccount, 3)
	pool.Add(req2new)
	pool.Add(req2new)
	require.EqualValues(t, 4, pool.refLUT.Size())
	require.EqualValues(t, 1, pool.reqsByAcountOrdered.Size())
	reqsInPoolForAccount, _ = pool.reqsByAcountOrdered.Get(agentID.String())
	require.Len(t, reqsInPoolForAccount, 4)

	// try to remove everything during iteration
	pool.Iterate(func(account string, entries []*OrderedPoolEntry[isc.OffLedgerRequest]) {
		for _, e := range entries {
			pool.Remove(e.req)
		}
	})
	require.EqualValues(t, 0, pool.refLUT.Size())
	require.EqualValues(t, 0, pool.reqsByAcountOrdered.Size())
}
