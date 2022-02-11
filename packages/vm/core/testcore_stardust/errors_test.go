package testcore

import (
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core"
	"github.com/iotaledger/wasp/packages/vm/core/errors"
	"github.com/iotaledger/wasp/packages/vm/runvm"
	"github.com/stretchr/testify/require"
	"testing"
)

/*
var errorContractName = "ErrorContract"
var errorContract = coreutil.NewContract(errorContractName, "error contract")

var funcRegisterErrors = coreutil.Func("register_errors")
var funcThrowErrors = coreutil.Func("throw_errors")

var testError uint16

var errorContractProcessor = errorContract.Processor(nil,
	funcRegisterErrors.WithHandler(func(ctx iscp.Sandbox) dict.Dict {
		testError = ctx.RegisterError("Test Error")

		return nil
	}),
	funcThrowErrors.WithHandler(func(ctx iscp.Sandbox) dict.Dict {
		panic(testError)
		return nil
	}),
)*/

func setupErrorsTest(t *testing.T) (*solo.Solo, *solo.Chain) {
	core.PrintWellKnownHnames()
	env := solo.New(t, &solo.InitOptions{AutoAdjustDustDeposit: true, Debug: true})
	chain, _, _ := env.NewChainExt(nil, 100_000, "chain1")

	chain.MustDepositIotasToL2(10_000_000, nil)
	defer chain.Log.Sync()
	chain.CheckChain()

	return env, chain
}

func TestErrorWithoutErrorMessage(t *testing.T) {
	_, chain := setupErrorsTest(t)

	req := solo.NewCallParams(errors.Contract.Name, errors.FuncRegisterError.Name).
		WithGasBudget(100_000)

	_, _, err := chain.PostRequestSyncTx(req, nil)

	require.ErrorIs(t, err, runvm.ErrUndefinedError)
}

func TestErrorWithErrorMessage(t *testing.T) {
	_, chain := setupErrorsTest(t)

	req := solo.NewCallParams(errors.Contract.Name, errors.FuncRegisterError.Name, errors.ParamErrorMessageFormat, "poof").
		WithGasBudget(100_000)

	_, _, err := chain.PostRequestSyncTx(req, nil)

	require.NoError(t, err)
}
