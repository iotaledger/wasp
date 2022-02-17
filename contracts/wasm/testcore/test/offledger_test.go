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

		accountBalance := ctx.Balance(ctx.Account())
		chainBalance := ctx.Balance(ctx.ChainAccount())
		originatorBalance := ctx.Balance(ctx.Originator())

		user := ctx.NewSoloAgent()
		ctx.Accounts(user)
		require.EqualValues(t, solo.Saldo, user.Balance())
		require.EqualValues(t, 0, ctx.Balance(user))

		// no deposit yet, so account is unverified

		f := testcore.ScFuncs.SetInt(ctx.OffLedger(user))
		f.Params.Name().SetValue("ppp")
		f.Params.IntValue().SetValue(314)
		f.Func.Post()
		require.Error(t, ctx.Err)
		require.Contains(t, ctx.Err.Error(), "gas budget exceeded")
		ctx.Accounts(user)

		require.EqualValues(t, accountBalance, ctx.Balance(ctx.Account()))
		require.EqualValues(t, chainBalance, ctx.Balance(ctx.ChainAccount()))
		require.EqualValues(t, originatorBalance, ctx.Balance(ctx.Originator()))
	})
}

func TestOffLedgerNoTransfer(t *testing.T) {
	run2(t, func(t *testing.T, w bool) {
		ctx := deployTestCore(t, w)

		accountBalance := ctx.Balance(ctx.Account())
		chainBalance := ctx.Balance(ctx.ChainAccount())
		originatorBalance := ctx.Balance(ctx.Originator())

		user := ctx.NewSoloAgent()
		ctx.Accounts(user)
		require.EqualValues(t, solo.Saldo, user.Balance())
		require.EqualValues(t, 0, ctx.Balance(user))

		ctx.Chain.MustDepositIotasToL2(1000+gasFee, user.Pair)
		ctx.Accounts(user)
		require.EqualValues(t, solo.Saldo-1000-gasFee, user.Balance())
		require.EqualValues(t, 1000, ctx.Balance(user))

		require.EqualValues(t, accountBalance, ctx.Balance(ctx.Account()))
		require.EqualValues(t, chainBalance+gasFee, ctx.Balance(ctx.ChainAccount()))
		require.EqualValues(t, originatorBalance, ctx.Balance(ctx.Originator()))

		// Look, Ma! No .TransferIotas() necessary when doing off-ledger request!
		// we're using setInt() here to be able to verify the state update was done
		f := testcore.ScFuncs.SetInt(ctx.OffLedger(user))
		f.Params.Name().SetValue("ppp")
		f.Params.IntValue().SetValue(314)
		f.Func.Post()
		require.NoError(t, ctx.Err)
		ctx.Accounts(user)

		require.EqualValues(t, solo.Saldo-1000-gasFee, user.Balance())
		require.EqualValues(t, 1000-gasFee, ctx.Balance(user))

		require.EqualValues(t, accountBalance, ctx.Balance(ctx.Account()))
		require.EqualValues(t, chainBalance+gasFee*2, ctx.Balance(ctx.ChainAccount()))
		require.EqualValues(t, originatorBalance, ctx.Balance(ctx.Originator()))

		// verify state update
		v := testcore.ScFuncs.GetInt(ctx)
		v.Params.Name().SetValue("ppp")
		v.Func.Call()
		require.NoError(t, ctx.Err)
		require.EqualValues(t, 314, v.Results.Values().GetInt64("ppp").Value())
	})
}

func TestOffLedgerTransferEnough(t *testing.T) {
	run2(t, func(t *testing.T, w bool) {
		ctx := deployTestCore(t, w)

		accountBalance := ctx.Balance(ctx.Account())
		chainBalance := ctx.Balance(ctx.ChainAccount())
		originatorBalance := ctx.Balance(ctx.Originator())

		user := ctx.NewSoloAgent()
		ctx.Accounts(user)
		require.EqualValues(t, solo.Saldo, user.Balance())
		require.EqualValues(t, 0, ctx.Balance(user))

		budget := uint64(1000)
		// deposit itself burns gas
		ctx.Chain.MustDepositIotasToL2(budget+gasFee, user.Pair)
		ctx.Accounts(user)
		require.EqualValues(t, solo.Saldo-budget-gasFee, user.Balance())
		require.EqualValues(t, budget, ctx.Balance(user))

		require.EqualValues(t, accountBalance, ctx.Balance(ctx.Account()))
		require.EqualValues(t, chainBalance+gasFee, ctx.Balance(ctx.ChainAccount()))
		require.EqualValues(t, originatorBalance, ctx.Balance(ctx.Originator()))

		// Transfer max amount of iotas that leaves enough for fee
		f := testcore.ScFuncs.SetInt(ctx.OffLedger(user))
		f.Params.Name().SetValue("ppp")
		f.Params.IntValue().SetValue(314)
		f.Func.TransferIotas(budget - gasFee).Post()
		require.NoError(t, ctx.Err)
		ctx.Accounts(user)

		require.EqualValues(t, solo.Saldo-budget-gasFee, user.Balance())
		require.EqualValues(t, budget-gasFee, ctx.Balance(user))

		require.EqualValues(t, accountBalance, ctx.Balance(ctx.Account()))
		require.EqualValues(t, chainBalance+gasFee*2, ctx.Balance(ctx.ChainAccount()))
		require.EqualValues(t, originatorBalance, ctx.Balance(ctx.Originator()))

		// verify state update
		v := testcore.ScFuncs.GetInt(ctx)
		v.Params.Name().SetValue("ppp")
		v.Func.Call()
		require.NoError(t, ctx.Err)
		require.EqualValues(t, 314, v.Results.Values().GetInt64("ppp").Value())
	})
}

func TestOffLedgerTransferNotEnough(t *testing.T) {
	run2(t, func(t *testing.T, w bool) {
		ctx := deployTestCore(t, w)

		accountBalance := ctx.Balance(ctx.Account())
		chainBalance := ctx.Balance(ctx.ChainAccount())
		originatorBalance := ctx.Balance(ctx.Originator())

		user := ctx.NewSoloAgent()
		ctx.Accounts(user)
		require.EqualValues(t, solo.Saldo, user.Balance())
		require.EqualValues(t, 0, ctx.Balance(user))

		budget := uint64(1000)
		ctx.Chain.MustDepositIotasToL2(budget+gasFee, user.Pair)
		ctx.Accounts(user)
		require.EqualValues(t, solo.Saldo-budget-gasFee, user.Balance())
		require.EqualValues(t, budget, ctx.Balance(user))

		require.EqualValues(t, accountBalance, ctx.Balance(ctx.Account()))
		require.EqualValues(t, chainBalance+gasFee, ctx.Balance(ctx.ChainAccount()))
		require.EqualValues(t, originatorBalance, ctx.Balance(ctx.Originator()))

		// Look, Ma! No .TransferIotas() necessary when doing off-ledger request!
		// we're using setInt() here to be able to verify the state update was done
		f := testcore.ScFuncs.SetInt(ctx.OffLedger(user))
		f.Params.Name().SetValue("ppp")
		f.Params.IntValue().SetValue(314)
		f.Func.TransferIotas(budget - gasFee + 1).Post()
		require.Error(t, ctx.Err)
		require.Contains(t, ctx.Err.Error(), "gas budget exceeded")
		ctx.Accounts(user)

		require.EqualValues(t, solo.Saldo-budget-gasFee, user.Balance())
		require.EqualValues(t, budget-gasFee, ctx.Balance(user))

		require.EqualValues(t, accountBalance, ctx.Balance(ctx.Account()))
		require.EqualValues(t, chainBalance+gasFee*2, ctx.Balance(ctx.ChainAccount()))
		require.EqualValues(t, originatorBalance, ctx.Balance(ctx.Originator()))

		// verify no state update
		v := testcore.ScFuncs.GetInt(ctx)
		v.Params.Name().SetValue("ppp")
		v.Func.Call()
		require.Error(t, ctx.Err)
		require.Contains(t, ctx.Err.Error(), "param 'ppp' not found")
	})
}
