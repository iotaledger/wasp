package test

import (
	"testing"

	"github.com/iotaledger/wasp/contracts/wasm/testcore/go/testcore"
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/stretchr/testify/require"
)

func TestDoNothing(t *testing.T) {
	run2(t, func(t *testing.T, w bool) {
		ctx := deployTestCore(t, w)
		bal := ctx.Balances()

		nop := testcore.ScFuncs.DoNothing(ctx)
		nop.Func.TransferBaseTokens(1 * isc.Mi).Post()
		require.NoError(t, ctx.Err)

		bal.Chain += ctx.GasFee
		bal.Originator += 1*isc.Mi - ctx.GasFee
		bal.VerifyBalances(t)
	})
}

func TestDoNothingUser(t *testing.T) {
	run2(t, func(t *testing.T, w bool) {
		ctx := deployTestCore(t, w)

		user := ctx.NewSoloAgent()
		bal := ctx.Balances(user)

		nop := testcore.ScFuncs.DoNothing(ctx.Sign(user))
		nop.Func.TransferBaseTokens(1 * isc.Mi).Post()
		require.NoError(t, ctx.Err)

		bal.Chain += ctx.GasFee
		bal.Add(user, 1*isc.Mi-ctx.GasFee)
		bal.VerifyBalances(t)
	})
}

func TestWithdrawToAddress(t *testing.T) {
	// TODO implement
	t.SkipNow()
	//	run2(t, func(t *testing.T, w bool) {
	//		ctx := deployTestCore(t, w)
	//
	//		user := ctx.NewSoloAgent()
	//		bal := ctx.Balances(user)
	//
	//		nop := testcore.ScFuncs.DoNothing(ctx.Sign(user))
	//		nop.Func.TransferIotas(1*isc.Mi).Post()
	//		require.NoError(t, ctx.Err)
	//
	//		bal.Chain += ctx.GasFee
	//		bal.Add(user, 1*isc.Mi-ctx.GasFee)
	//		bal.VerifyBalances(t)
	//
	//		// TODO implement
	//		// send entire contract balance back to user
	//		xfer := testcore.ScFuncs.SendToAddress(ctx.Sign(ctx.Originator()))
	//		xfer.Params.Address().SetValue(user.ScAddress())
	//		xfer.Func.Post()
	//		require.NoError(t, ctx.Err)
	//
	//		t.Logf("dump accounts:\n%s", ctx.Chain.DumpAccounts())
	//		require.EqualValues(t, solo.Saldo-42+42+1, user.Balance())
	//		require.EqualValues(t, 0, ctx.Balance(ctx.Account()))
	//		require.EqualValues(t, 0, ctx.Balance(ctx.Originator()))
	//		require.EqualValues(t, 0, ctx.Balance(user))
	//		originatorBalanceReducedBy(ctx, w, 2+1)
	//		chainAccountBalances(ctx, w, 2, 2)
	//	})
}

func TestDoPanicUser(t *testing.T) {
	run2(t, func(t *testing.T, w bool) {
		ctx := deployTestCore(t, w)

		user := ctx.NewSoloAgent()
		bal := ctx.Balances(user)
		userL1 := user.Balance()

		f := testcore.ScFuncs.TestPanicFullEP(ctx.Sign(user))
		f.Func.TransferBaseTokens(1 * isc.Mi).Post()
		require.Error(t, ctx.Err)
		require.EqualValues(t, userL1-1*isc.Mi, user.Balance())

		bal.Chain += ctx.GasFee
		bal.Add(user, 1*isc.Mi-ctx.GasFee)
		bal.VerifyBalances(t)
	})
}

func TestDoPanicUserFeeless(t *testing.T) {
	// TODO implement
	t.SkipNow()
	//run2(t, func(t *testing.T, w bool) {
	//	ctx := deployTestCore(t, w)
	//	user := ctx.NewSoloAgent()
	//
	//	t.Logf("dump accounts:\n%s", ctx.Chain.DumpAccounts())
	//	require.EqualValues(t, utxodb.FundsFromFaucetAmount, user.Balance())
	//	require.EqualValues(t, 0, ctx.Balance(ctx.Account()))
	//	require.EqualValues(t, 0, ctx.Balance(ctx.Originator()))
	//	require.EqualValues(t, 0, ctx.Balance(user))
	//	originatorBalanceReducedBy(ctx, w, 2)
	//	chainAccountBalances(ctx, w, 2, 2)
	//
	//	f := testcore.ScFuncs.TestPanicFullEP(ctx.Sign(user))
	//	f.Func.TransferIotas(1 * isc.Mi).Post()
	//	require.Error(t, ctx.Err)
	//
	//	t.Logf("dump accounts:\n%s", ctx.Chain.DumpAccounts())
	//	require.EqualValues(t, utxodb.FundsFromFaucetAmount, user.Balance())
	//	require.EqualValues(t, 0, ctx.Balance(ctx.Account()))
	//	require.EqualValues(t, 0, ctx.Balance(ctx.Originator()))
	//	require.EqualValues(t, 0, ctx.Balance(user))
	//	originatorBalanceReducedBy(ctx, w, 2)
	//	chainAccountBalances(ctx, w, 2, 2)
	//
	//	withdraw(t, ctx, user)
	//
	//	t.Logf("dump accounts:\n%s", ctx.Chain.DumpAccounts())
	//	require.EqualValues(t, utxodb.FundsFromFaucetAmount-1, user.Balance())
	//	require.EqualValues(t, 0, ctx.Balance(ctx.Account()))
	//	require.EqualValues(t, 0, ctx.Balance(ctx.Originator()))
	//	require.EqualValues(t, 0, ctx.Balance(user))
	//	originatorBalanceReducedBy(ctx, w, 2)
	//	chainAccountBalances(ctx, w, 3, 3)
	//})
}

func TestDoPanicUserFee(t *testing.T) {
	// TODO implement
	t.SkipNow()
	//run2(t, func(t *testing.T, w bool) {
	//	ctx := deployTestCore(t, w)
	//	user := ctx.NewSoloAgent()
	//
	//	t.Logf("dump accounts:\n%s", ctx.Chain.DumpAccounts())
	//	require.EqualValues(t, utxodb.FundsFromFaucetAmount, user.Balance())
	//	require.EqualValues(t, 0, ctx.Balance(ctx.Account()))
	//	require.EqualValues(t, 0, ctx.Balance(ctx.Originator()))
	//	require.EqualValues(t, 0, ctx.Balance(user))
	//	originatorBalanceReducedBy(ctx, w, 2)
	//	chainAccountBalances(ctx, w, 2, 2)
	//
	//	setOwnerFee(t, ctx, 10)
	//
	//	t.Logf("dump accounts:\n%s", ctx.Chain.DumpAccounts())
	//	require.EqualValues(t, utxodb.FundsFromFaucetAmount, user.Balance())
	//	require.EqualValues(t, 0, ctx.Balance(ctx.Account()))
	//	require.EqualValues(t, 0, ctx.Balance(ctx.Originator()))
	//	require.EqualValues(t, 0, ctx.Balance(user))
	//	originatorBalanceReducedBy(ctx, w, 3)
	//	chainAccountBalances(ctx, w, 3, 3)
	//
	//	f := testcore.ScFuncs.TestPanicFullEP(ctx.Sign(user))
	//	f.Func.TransferIotas(1*isc.Mi).Post()
	//	require.Error(t, ctx.Err)
	//
	//	t.Logf("dump accounts:\n%s", ctx.Chain.DumpAccounts())
	//	require.EqualValues(t, utxodb.FundsFromFaucetAmount-10, user.Balance())
	//	require.EqualValues(t, 0, ctx.Balance(ctx.Account()))
	//	require.EqualValues(t, 0, ctx.Balance(ctx.Originator()))
	//	require.EqualValues(t, 0, ctx.Balance(user))
	//	originatorBalanceReducedBy(ctx, w, 3)
	//	chainAccountBalances(ctx, w, 3+10, 3+10)
	//})
}

func TestRequestToView(t *testing.T) {
	run2(t, func(t *testing.T, w bool) {
		ctx := deployTestCore(t, w)
		user := ctx.NewSoloAgent()
		userL1 := user.Balance()
		bal := ctx.Balances(user)

		// SoloContext prevents Sign()/Post() to a view
		// Therefore we cannot simply do the following:
		// f := testcore.ScFuncs.JustView(ctx.Sign(user))
		// f.Func.TransferIotas(1*isc.Mi).Post()
		// require.Error(t, ctx.Err)

		// sending request to the view entry point should
		// return an error and leave tokens in L2 minus gas fee
		req := solo.NewCallParams(testcore.ScName, testcore.ViewJustView)
		_, ctx.Err = ctx.Chain.PostRequestSync(req.AddBaseTokens(1*isc.Mi), user.Pair)
		require.Error(t, ctx.Err)
		require.EqualValues(t, userL1-1*isc.Mi, user.Balance())
		ctx.UpdateGasFees()

		bal.Chain += ctx.GasFee
		bal.Add(user, 1*isc.Mi-ctx.GasFee)
		bal.VerifyBalances(t)
	})
}
