package test

import (
	"testing"

	"github.com/iotaledger/wasp/contracts/wasm/testcore/go/testcore"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/stretchr/testify/require"
)

const gasFee = uint64(100)

func TestOffLedgerFailNoAccount(t *testing.T) {
	run2(t, func(t *testing.T, w bool) {
		ctx := deployTestCore(t, w)

		user := ctx.NewSoloAgent()
		require.EqualValues(t, solo.Saldo, user.Balance())
		require.EqualValues(t, 0, ctx.Balance(user))
		bal := ctx.Balances(user)

		// no deposit yet, so account is unverified

		f := testcore.ScFuncs.SetInt(ctx.OffLedger(user))
		f.Params.Name().SetValue("ppp")
		f.Params.IntValue().SetValue(314)
		f.Func.Post()
		require.Error(t, ctx.Err)
		require.Contains(t, ctx.Err.Error(), "gas budget exceeded")
		bal.VerifyBalances(t)
	})
}

func TestOffLedgerNoTransfer(t *testing.T) {
	run2(t, func(t *testing.T, w bool) {
		ctx := deployTestCore(t, w)

		user := ctx.NewSoloAgent()
		require.EqualValues(t, solo.Saldo, user.Balance())
		require.EqualValues(t, 0, ctx.Balance(user))
		bal := ctx.Balances(user)

		budget := uint64(5000)
		ctx.Chain.MustDepositIotasToL2(budget+gasFee, user.Pair)
		require.EqualValues(t, solo.Saldo-budget-gasFee, user.Balance())

		bal.Chain += gasFee
		bal.Add(user, budget)
		bal.VerifyBalances(t)

		// Look, Ma! No .TransferIotas() necessary when doing off-ledger request!
		// we're using setInt() here to be able to verify the state update was done
		f := testcore.ScFuncs.SetInt(ctx.OffLedger(user))
		f.Params.Name().SetValue("ppp")
		f.Params.IntValue().SetValue(314)
		f.Func.Post()
		require.NoError(t, ctx.Err)

		require.EqualValues(t, solo.Saldo-budget-gasFee, user.Balance())

		bal.Chain += ctx.GasFee
		bal.Add(user, -ctx.GasFee)
		bal.VerifyBalances(t)

		// verify state update
		v := testcore.ScFuncs.GetInt(ctx)
		v.Params.Name().SetValue("ppp")
		v.Func.Call()
		require.NoError(t, ctx.Err)
		require.EqualValues(t, 314, v.Results.Values().GetInt64("ppp").Value())
	})
}

func TestOffLedgerTransferWhenEnoughBudget(t *testing.T) {
	run2(t, func(t *testing.T, w bool) {
		ctx := deployTestCore(t, w)

		user := ctx.NewSoloAgent()
		require.EqualValues(t, solo.Saldo, user.Balance())
		require.EqualValues(t, 0, ctx.Balance(user))
		bal := ctx.Balances(user)

		budget := uint64(9999)
		ctx.Chain.MustDepositIotasToL2(budget+gasFee, user.Pair)
		require.EqualValues(t, solo.Saldo-budget-gasFee, user.Balance())

		bal.Chain += gasFee
		bal.Add(user, budget)
		bal.VerifyBalances(t)

		// Allow 1000 iotas to be transferred. there's enough budget
		// note that SetInt() will not try to grab them
		f := testcore.ScFuncs.SetInt(ctx.OffLedger(user))
		f.Params.Name().SetValue("ppp")
		f.Params.IntValue().SetValue(314)
		f.Func.TransferIotas(1000).Post()
		require.NoError(t, ctx.Err)
		ctx.Balances(user)

		require.EqualValues(t, solo.Saldo-budget-gasFee, user.Balance())

		bal.Chain += ctx.GasFee
		bal.Add(user, -ctx.GasFee)
		bal.VerifyBalances(t)

		// verify state update
		v := testcore.ScFuncs.GetInt(ctx)
		v.Params.Name().SetValue("ppp")
		v.Func.Call()
		require.NoError(t, ctx.Err)
		require.EqualValues(t, 314, v.Results.Values().GetInt64("ppp").Value())
	})
}

func TestOffLedgerTransferWhenNotEnoughBudget(t *testing.T) {
	run2(t, func(t *testing.T, w bool) {
		ctx := deployTestCore(t, w)

		user := ctx.NewSoloAgent()
		require.EqualValues(t, solo.Saldo, user.Balance())
		require.EqualValues(t, 0, ctx.Balance(user))
		bal := ctx.Balances(user)

		budget := uint64(1000)
		ctx.Chain.MustDepositIotasToL2(budget+gasFee, user.Pair)
		require.EqualValues(t, solo.Saldo-budget-gasFee, user.Balance())

		bal.Chain += gasFee
		bal.Add(user, budget)
		bal.VerifyBalances(t)

		// Allow 1000 iotas to be transferred. there's not enough budget
		// note that SetInt() will not try to grab them
		f := testcore.ScFuncs.SetInt(ctx.OffLedger(user))
		f.Params.Name().SetValue("ppp")
		f.Params.IntValue().SetValue(314)
		f.Func.TransferIotas(budget - gasFee + 1).Post()
		require.Error(t, ctx.Err)
		require.Contains(t, ctx.Err.Error(), "gas budget exceeded")

		require.EqualValues(t, solo.Saldo-budget-gasFee, user.Balance())

		bal.Chain += ctx.GasFee
		bal.Add(user, -ctx.GasFee)
		bal.VerifyBalances(t)

		// verify no state update
		v := testcore.ScFuncs.GetInt(ctx)
		v.Params.Name().SetValue("ppp")
		v.Func.Call()
		require.Error(t, ctx.Err)
		require.Contains(t, ctx.Err.Error(), "param 'ppp' not found")
	})
}
