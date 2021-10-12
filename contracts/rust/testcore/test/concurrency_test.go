package test

import (
	"testing"
	"time"

	"github.com/iotaledger/wasp/contracts/rust/testcore"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/wasmsolo"
	"github.com/stretchr/testify/require"
)

func TestCounter(t *testing.T) {
	run2(t, func(t *testing.T, w bool) {
		ctx := setupTest(t, w)

		f := testcore.ScFuncs.IncCounter(ctx)
		f.Func.TransferIotas(1)
		for i := 0; i < 33; i++ {
			f.Func.Post()
			require.NoError(t, ctx.Err)
		}

		v := testcore.ScFuncs.GetCounter(ctx)
		v.Func.Call()
		require.NoError(t, ctx.Err)
		require.EqualValues(t, 33, v.Results.Counter().Value())
	})
}

func TestSynchronous(t *testing.T) {
	run2(t, func(t *testing.T, w bool) {
		ctx := setupTest(t, w)

		f := testcore.ScFuncs.IncCounter(ctx)
		f.Func.TransferIotas(1)

		repeats := []int{300, 100, 100, 100, 200, 100, 100}
		sum := 0
		for _, i := range repeats {
			sum += i
		}

		for _, n := range repeats {
			for i := 0; i < n; i++ {
				ctx.EnqueueRequest()
				f.Func.Post()
				require.NoError(t, ctx.Err)
			}
		}
		reqs := sum + 2
		if w {
			reqs++
		}
		require.True(t, ctx.WaitForPendingRequests(-reqs, 20*time.Second))

		v := testcore.ScFuncs.GetCounter(ctx)
		v.Func.Call()
		require.NoError(t, ctx.Err)
		require.EqualValues(t, sum, v.Results.Counter().Value())

		require.EqualValues(t, sum, ctx.Balance(ctx.Account()))
		chainAccountBalances(ctx, w, 2, uint64(2+sum))
	})
}

func TestConcurrency(t *testing.T) {
	run2(t, func(t *testing.T, w bool) {
		ctx := setupTest(t, w)

		// note that because SoloContext is not thread-safe we cannot use
		// the following in parallel go-routines
		f := testcore.ScFuncs.IncCounter(ctx)
		f.Func.TransferIotas(1)

		req := solo.NewCallParams(testcore.ScName, testcore.FuncIncCounter).
			WithIotas(1)

		repeats := []int{300, 100, 100, 100, 200, 100, 100}
		sum := 0
		for _, i := range repeats {
			sum += i
		}

		chain := ctx.Chain
		for r, n := range repeats {
			go func(_, n int) {
				for i := 0; i < n; i++ {
					tx, _, err := chain.RequestFromParamsToLedger(req, nil)
					require.NoError(t, err)
					chain.Env.EnqueueRequests(tx)
				}
			}(r, n)
		}
		require.True(t, ctx.WaitForPendingRequests(sum, 20*time.Second))

		v := testcore.ScFuncs.GetCounter(ctx)
		v.Func.Call()
		require.NoError(t, ctx.Err)
		require.EqualValues(t, sum, v.Results.Counter().Value())

		require.EqualValues(t, sum, ctx.Balance(ctx.Account()))
		chainAccountBalances(ctx, w, 2, uint64(2+sum))
	})
}

func TestConcurrency2(t *testing.T) {
	run2(t, func(t *testing.T, w bool) {
		ctx := setupTest(t, w)

		// note that because SoloContext is not thread-safe we cannot use
		// the following in parallel go-routines
		f := testcore.ScFuncs.IncCounter(ctx)
		f.Func.TransferIotas(1)

		req := solo.NewCallParams(testcore.ScName, testcore.FuncIncCounter).
			WithIotas(1)

		repeats := []int{300, 100, 100, 100, 200, 100, 100}
		sum := 0
		for _, i := range repeats {
			sum += i
		}

		chain := ctx.Chain
		users := make([]*wasmsolo.SoloAgent, len(repeats))
		for r, n := range repeats {
			go func(r, n int) {
				users[r] = ctx.NewSoloAgent()
				for i := 0; i < n; i++ {
					tx, _, err := chain.RequestFromParamsToLedger(req, users[r].Pair)
					require.NoError(t, err)
					chain.Env.EnqueueRequests(tx)
				}
			}(r, n)
		}

		require.True(t, ctx.WaitForPendingRequests(sum, 20*time.Second))

		v := testcore.ScFuncs.GetCounter(ctx)
		v.Func.Call()
		require.NoError(t, ctx.Err)
		require.EqualValues(t, sum, v.Results.Counter().Value())

		for i, user := range users {
			require.EqualValues(t, solo.Saldo-repeats[i], user.Balance())
			require.EqualValues(t, 0, ctx.Balance(user))
		}

		require.EqualValues(t, sum, ctx.Balance(ctx.Account()))
		require.EqualValues(t, sum, ctx.Balance(ctx.Account()))
		chainAccountBalances(ctx, w, 2, uint64(2+sum))
	})
}
