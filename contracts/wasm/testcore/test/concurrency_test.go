package test

import (
	"errors"
	"testing"
	"time"

	"github.com/iotaledger/wasp/contracts/wasm/testcore/go/testcore"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/wasmsolo"
	"github.com/stretchr/testify/require"
)

func TestCounter(t *testing.T) {
	run2(t, func(t *testing.T, w bool) {
		ctx := deployTestCore(t, w)

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
		// TODO fails with 999 instead of 1000 at WaitForPendingRequests
		if *wasmsolo.GoDebug || *wasmsolo.GoWasmEdge {
			t.SkipNow()
		}
		ctx := deployTestCore(t, w)

		f := testcore.ScFuncs.IncCounter(ctx)
		f.Func.TransferIotas(1)

		repeats := []int{300, 100, 100, 100, 200, 100, 100}
		if wasmsolo.SoloDebug {
			for i := range repeats {
				repeats[i] /= 10
			}
		}

		sum := 0
		for _, n := range repeats {
			sum += n
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
		require.True(t, ctx.WaitForPendingRequests(-reqs, 180*time.Second))

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
		ctx := deployTestCore(t, w)

		// note that because SoloContext is not thread-safe we cannot use
		// the following in parallel go-routines
		f := testcore.ScFuncs.IncCounter(ctx)
		f.Func.TransferIotas(1)

		req := solo.NewCallParams(testcore.ScName, testcore.FuncIncCounter).
			WithIotas(1)

		repeats := []int{300, 100, 100, 100, 200, 100, 100}
		if wasmsolo.SoloDebug {
			for i := range repeats {
				repeats[i] /= 10
			}
		}

		sum := 0
		for _, n := range repeats {
			sum += n
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
		require.True(t, ctx.WaitForPendingRequests(sum, 180*time.Second))

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
		ctx := deployTestCore(t, w)

		// note that because SoloContext is not thread-safe we cannot use
		// the following in parallel go-routines
		f := testcore.ScFuncs.IncCounter(ctx)
		f.Func.TransferIotas(1)

		req := solo.NewCallParams(testcore.ScName, testcore.FuncIncCounter).
			WithIotas(1)

		repeats := []int{300, 100, 100, 100, 200, 100, 100}
		if wasmsolo.SoloDebug {
			for i := range repeats {
				repeats[i] /= 10
			}
		}

		sum := 0
		for _, n := range repeats {
			sum += n
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

		require.True(t, ctx.WaitForPendingRequests(sum, 180*time.Second))

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

func TestViewConcurrency(t *testing.T) {
	run2(t, func(t *testing.T, w bool) {
		ctx := deployTestCore(t, false)

		f := testcore.ScFuncs.IncCounter(ctx)
		f.Func.TransferIotas(1).Post()

		times := 2000
		if wasmsolo.SoloDebug {
			times /= 10
		}

		channels := make(chan error, times)
		chain := ctx.Chain
		for i := 0; i < times; i++ {
			go func() {
				res, err := chain.CallView(testcore.ScName, testcore.ViewGetCounter)
				if err != nil {
					channels <- err
					return
				}
				v, err := codec.DecodeInt64(res.MustGet("counter"))
				if err == nil && v != 1 {
					err = errors.New("v != 1")
				}
				channels <- err
			}()
		}

		for i := 0; i < times; i++ {
			err := <-channels
			require.NoError(t, err)
		}

		v := testcore.ScFuncs.GetCounter(ctx)
		v.Func.Call()
		require.NoError(t, ctx.Err)
		require.EqualValues(t, 1, v.Results.Counter().Value())
	})
}
