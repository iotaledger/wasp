package test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/contracts/wasm/testcore/go/testcore"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/wasmvm/wasmsolo"
)

func TestCounter(t *testing.T) {
	run2(t, func(t *testing.T, w bool) {
		ctx := deployTestCore(t, w)

		f := testcore.ScFuncs.IncCounter(ctx)
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
		ctx := deployTestCore(t, w)

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

		ctx.WaitForPendingRequestsMark()

		f := testcore.ScFuncs.IncCounter(ctx)
		for _, n := range repeats {
			for i := 0; i < n; i++ {
				ctx.EnqueueRequest()
				f.Func.Post()
				require.NoError(t, ctx.Err)
			}
		}

		require.True(t, ctx.WaitForPendingRequests(sum, 180*time.Second))

		v := testcore.ScFuncs.GetCounter(ctx)
		v.Func.Call()
		require.NoError(t, ctx.Err)
		require.EqualValues(t, sum, v.Results.Counter().Value())

		// require.EqualValues(t, sum, ctx.Balance(ctx.Account()))
		// chainAccountBalances(ctx, w, 2, uint64(2+sum))
	})
}

func TestConcurrency(t *testing.T) {
	run2(t, func(t *testing.T, w bool) {
		ctx := deployTestCore(t, w)

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

		ctx.WaitForPendingRequestsMark()

		f := testcore.ScFuncs.IncCounter(ctx)
		for r, n := range repeats {
			go func(_, n int) {
				for i := 0; i < n; i++ {
					ctx.EnqueueRequest()
					f.Func.Post()
					require.NoError(t, ctx.Err)
				}
			}(r, n)
		}

		require.True(t, ctx.WaitForPendingRequests(sum, 180*time.Second))

		v := testcore.ScFuncs.GetCounter(ctx)
		v.Func.Call()
		require.NoError(t, ctx.Err)
		require.EqualValues(t, sum, v.Results.Counter().Value())

		// require.EqualValues(t, sum, ctx.Balance(ctx.Account()))
		// chainAccountBalances(ctx, w, 2, uint64(2+sum))
	})
}

func TestConcurrency2(t *testing.T) {
	run2(t, func(t *testing.T, w bool) {
		ctx := deployTestCore(t, w)

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

		users := make([]*wasmsolo.SoloAgent, len(repeats))
		funcs := make([]*testcore.IncCounterCall, len(repeats))
		for r := range repeats {
			users[r] = ctx.NewSoloAgent()
			funcs[r] = testcore.ScFuncs.IncCounter(ctx.Sign(users[r]))
		}

		ctx.WaitForPendingRequestsMark()

		for r, n := range repeats {
			go func(r, n int) {
				for i := 0; i < n; i++ {
					ctx.EnqueueRequest()
					funcs[r].Func.Post()
					require.NoError(t, ctx.Err)
				}
			}(r, n)
		}

		require.True(t, ctx.WaitForPendingRequests(sum, 180*time.Second))

		v := testcore.ScFuncs.GetCounter(ctx)
		v.Func.Call()
		require.NoError(t, ctx.Err)
		require.EqualValues(t, sum, v.Results.Counter().Value())

		//for i, user := range users {
		//	require.EqualValues(t, utxodb.FundsFromFaucetAmount-uint64(repeats[i]), user.Balance())
		//	require.EqualValues(t, 0, ctx.Balance(user))
		//}

		// require.EqualValues(t, sum, ctx.Balance(ctx.Account()))
		// require.EqualValues(t, sum, ctx.Balance(ctx.Account()))
		// chainAccountBalances(ctx, w, 2, uint64(2+sum))
	})
}

func TestViewConcurrency(t *testing.T) {
	run2(t, func(t *testing.T, w bool) {
		ctx := deployTestCore(t, false)

		f := testcore.ScFuncs.IncCounter(ctx)
		f.Func.Post()

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
