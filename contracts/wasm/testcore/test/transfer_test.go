package test

import (
	"testing"

	"github.com/iotaledger/wasp/contracts/wasm/testcore/go/testcore"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/stretchr/testify/require"
)

func TestDoNothing(t *testing.T) {
	run2(t, func(t *testing.T, w bool) {
		ctx := deployTestCore(t, w)

		nop := testcore.ScFuncs.DoNothing(ctx)
		nop.Func.TransferIotas(42).Post()
		require.NoError(t, ctx.Err)

		t.Logf("dump accounts:\n%s", ctx.Chain.DumpAccounts())
		require.EqualValues(t, 42, ctx.Balance(ctx.Account()))
		require.EqualValues(t, 0, ctx.Balance(ctx.Originator()))
		originatorBalanceReducedBy(ctx, w, 2+42)
		chainAccountBalances(ctx, w, 2, 2+42)
	})
}

func TestDoNothingUser(t *testing.T) {
	run2(t, func(t *testing.T, w bool) {
		ctx := deployTestCore(t, w)

		user := ctx.NewSoloAgent()
		nop := testcore.ScFuncs.DoNothing(ctx.Sign(user))
		nop.Func.TransferIotas(42).Post()
		require.NoError(t, ctx.Err)

		t.Logf("dump accounts:\n%s", ctx.Chain.DumpAccounts())
		require.EqualValues(t, solo.Saldo-42, user.Balance())
		require.EqualValues(t, 42, ctx.Balance(ctx.Account()))
		require.EqualValues(t, 0, ctx.Balance(ctx.Originator()))
		require.EqualValues(t, 0, ctx.Balance(user))
		originatorBalanceReducedBy(ctx, w, 2)
		chainAccountBalances(ctx, w, 2, 2+42)
	})
}

func TestWithdrawToAddress(t *testing.T) {
	run2(t, func(t *testing.T, w bool) {
		ctx := deployTestCore(t, w)
		user := ctx.NewSoloAgent()

		nop := testcore.ScFuncs.DoNothing(ctx.Sign(user))
		nop.Func.TransferIotas(42).Post()
		require.NoError(t, ctx.Err)

		t.Logf("dump accounts:\n%s", ctx.Chain.DumpAccounts())
		require.EqualValues(t, solo.Saldo-42, user.Balance())
		require.EqualValues(t, 42, ctx.Balance(ctx.Account()))
		require.EqualValues(t, 0, ctx.Balance(ctx.Originator()))
		require.EqualValues(t, 0, ctx.Balance(user))
		originatorBalanceReducedBy(ctx, w, 2)
		chainAccountBalances(ctx, w, 2, 2+42)

		// send entire contract balance back to user
		// note that that includes the token that we transfer here
		xfer := testcore.ScFuncs.SendToAddress(ctx.Sign(ctx.Originator()))
		xfer.Params.Address().SetValue(user.ScAddress())
		xfer.Func.TransferIotas(1).Post()
		require.NoError(t, ctx.Err)

		t.Logf("dump accounts:\n%s", ctx.Chain.DumpAccounts())
		require.EqualValues(t, solo.Saldo-42+42+1, user.Balance())
		require.EqualValues(t, 0, ctx.Balance(ctx.Account()))
		require.EqualValues(t, 0, ctx.Balance(ctx.Originator()))
		require.EqualValues(t, 0, ctx.Balance(user))
		originatorBalanceReducedBy(ctx, w, 2+1)
		chainAccountBalances(ctx, w, 2, 2)
	})
}

func TestDoPanicUser(t *testing.T) {
	run2(t, func(t *testing.T, w bool) {
		ctx := deployTestCore(t, w)
		user := ctx.NewSoloAgent()

		t.Logf("dump accounts:\n%s", ctx.Chain.DumpAccounts())
		require.EqualValues(t, solo.Saldo, user.Balance())
		require.EqualValues(t, 0, ctx.Balance(ctx.Account()))
		require.EqualValues(t, 0, ctx.Balance(ctx.Originator()))
		require.EqualValues(t, 0, ctx.Balance(user))
		originatorBalanceReducedBy(ctx, w, 2)
		chainAccountBalances(ctx, w, 2, 2)

		f := testcore.ScFuncs.TestPanicFullEP(ctx.Sign(user))
		f.Func.TransferIotas(42).Post()
		require.Error(t, ctx.Err)

		t.Logf("dump accounts:\n%s", ctx.Chain.DumpAccounts())
		require.EqualValues(t, solo.Saldo, user.Balance())
		require.EqualValues(t, 0, ctx.Balance(ctx.Account()))
		require.EqualValues(t, 0, ctx.Balance(ctx.Originator()))
		require.EqualValues(t, 0, ctx.Balance(user))
		originatorBalanceReducedBy(ctx, w, 2)
		chainAccountBalances(ctx, w, 2, 2)
	})
}

func TestDoPanicUserFeeless(t *testing.T) {
	run2(t, func(t *testing.T, w bool) {
		ctx := deployTestCore(t, w)
		user := ctx.NewSoloAgent()

		t.Logf("dump accounts:\n%s", ctx.Chain.DumpAccounts())
		require.EqualValues(t, solo.Saldo, user.Balance())
		require.EqualValues(t, 0, ctx.Balance(ctx.Account()))
		require.EqualValues(t, 0, ctx.Balance(ctx.Originator()))
		require.EqualValues(t, 0, ctx.Balance(user))
		originatorBalanceReducedBy(ctx, w, 2)
		chainAccountBalances(ctx, w, 2, 2)

		f := testcore.ScFuncs.TestPanicFullEP(ctx.Sign(user))
		f.Func.TransferIotas(42).Post()
		require.Error(t, ctx.Err)

		t.Logf("dump accounts:\n%s", ctx.Chain.DumpAccounts())
		require.EqualValues(t, solo.Saldo, user.Balance())
		require.EqualValues(t, 0, ctx.Balance(ctx.Account()))
		require.EqualValues(t, 0, ctx.Balance(ctx.Originator()))
		require.EqualValues(t, 0, ctx.Balance(user))
		originatorBalanceReducedBy(ctx, w, 2)
		chainAccountBalances(ctx, w, 2, 2)

		withdraw(t, ctx, user)

		t.Logf("dump accounts:\n%s", ctx.Chain.DumpAccounts())
		require.EqualValues(t, solo.Saldo-1, user.Balance())
		require.EqualValues(t, 0, ctx.Balance(ctx.Account()))
		require.EqualValues(t, 0, ctx.Balance(ctx.Originator()))
		require.EqualValues(t, 0, ctx.Balance(user))
		originatorBalanceReducedBy(ctx, w, 2)
		chainAccountBalances(ctx, w, 3, 3)
	})
}

func TestDoPanicUserFee(t *testing.T) {
	run2(t, func(t *testing.T, w bool) {
		ctx := deployTestCore(t, w)
		user := ctx.NewSoloAgent()

		t.Logf("dump accounts:\n%s", ctx.Chain.DumpAccounts())
		require.EqualValues(t, solo.Saldo, user.Balance())
		require.EqualValues(t, 0, ctx.Balance(ctx.Account()))
		require.EqualValues(t, 0, ctx.Balance(ctx.Originator()))
		require.EqualValues(t, 0, ctx.Balance(user))
		originatorBalanceReducedBy(ctx, w, 2)
		chainAccountBalances(ctx, w, 2, 2)

		setOwnerFee(t, ctx, 10)

		t.Logf("dump accounts:\n%s", ctx.Chain.DumpAccounts())
		require.EqualValues(t, solo.Saldo, user.Balance())
		require.EqualValues(t, 0, ctx.Balance(ctx.Account()))
		require.EqualValues(t, 0, ctx.Balance(ctx.Originator()))
		require.EqualValues(t, 0, ctx.Balance(user))
		originatorBalanceReducedBy(ctx, w, 3)
		chainAccountBalances(ctx, w, 3, 3)

		f := testcore.ScFuncs.TestPanicFullEP(ctx.Sign(user))
		f.Func.TransferIotas(42).Post()
		require.Error(t, ctx.Err)

		t.Logf("dump accounts:\n%s", ctx.Chain.DumpAccounts())
		require.EqualValues(t, solo.Saldo-10, user.Balance())
		require.EqualValues(t, 0, ctx.Balance(ctx.Account()))
		require.EqualValues(t, 0, ctx.Balance(ctx.Originator()))
		require.EqualValues(t, 0, ctx.Balance(user))
		originatorBalanceReducedBy(ctx, w, 3)
		chainAccountBalances(ctx, w, 3+10, 3+10)
	})
}

func TestRequestToView(t *testing.T) {
	run2(t, func(t *testing.T, w bool) {
		ctx := deployTestCore(t, w)
		user := ctx.NewSoloAgent()

		t.Logf("dump accounts:\n%s", ctx.Chain.DumpAccounts())
		require.EqualValues(t, solo.Saldo, user.Balance())
		require.EqualValues(t, 0, ctx.Balance(ctx.Account()))
		require.EqualValues(t, 0, ctx.Balance(ctx.Originator()))
		require.EqualValues(t, 0, ctx.Balance(user))
		originatorBalanceReducedBy(ctx, w, 2)
		chainAccountBalances(ctx, w, 2, 2)

		// SoloContext disallows Sign()/Post() to a view
		// f := testcore.ScFuncs.JustView(ctx.Sign(user))
		// f.Func.TransferIotas(42).Post()
		// require.Error(t, ctx.Err)

		// sending request to the view entry point should
		// return an error and invoke fallback for tokens
		req := solo.NewCallParams(testcore.ScName, testcore.ViewJustView)
		_, ctx.Err = ctx.Chain.PostRequestSync(req.WithIotas(42), user.Pair)
		require.Error(t, ctx.Err)

		t.Logf("dump accounts:\n%s", ctx.Chain.DumpAccounts())
		require.EqualValues(t, solo.Saldo, user.Balance())
		require.EqualValues(t, 0, ctx.Balance(ctx.Account()))
		require.EqualValues(t, 0, ctx.Balance(ctx.Originator()))
		require.EqualValues(t, 0, ctx.Balance(user))
		originatorBalanceReducedBy(ctx, w, 2)
		chainAccountBalances(ctx, w, 2, 2)
	})
}
