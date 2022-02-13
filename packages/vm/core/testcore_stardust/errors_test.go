package testcore

import (
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm/core"
	"github.com/iotaledger/wasp/packages/vm/core/errors"
	"github.com/iotaledger/wasp/packages/vm/core/errors/commonerrors"
	"github.com/iotaledger/wasp/packages/vm/vmcontext"
	"github.com/iotaledger/wasp/packages/vm/vmerrors"
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

func setupErrorsTestWithoutFunds(t *testing.T) (*solo.Solo, *solo.Chain) {
	core.PrintWellKnownHnames()
	env := solo.New(t, &solo.InitOptions{AutoAdjustDustDeposit: true, Debug: true})
	chain, _, _ := env.NewChainExt(nil, 1, "chain1")

	chain.MustDepositIotasToL2(1, nil)
	defer chain.Log.Sync()
	chain.CheckChain()

	return env, chain
}

// Panics and returned errors will eventually land into the error handling hook.
// Typical xerror/error types will be wrapped into a vmerrors.Error type (Err ErrUntypedError)
// Panicked vmerrors will be stored as is.
// The first test validates a typed vmerror Error (Not enough Gas)
// The second test validates the wrapped generic ErrUntypedError

func TestErrorWithCustomError(t *testing.T) {
	_, chain := setupErrorsTestWithoutFunds(t)

	req := solo.NewCallParams(errors.Contract.Name, errors.FuncRegisterError.Name).
		WithGasBudget(1)

	_, _, err := chain.PostRequestSyncTx(req, nil)

	testError := &vmerrors.Error{}
	require.ErrorAs(t, err, &testError)

	typedError := err.(*vmerrors.Error)
	require.Equal(t, typedError.Definition(), *vmcontext.ErrGasBudgetDetail)
}

// This test does not supply the required kv pair 'ParamErrorMessageFormat' which makes the kvdecoder fail with an xerror
func TestPanicDueMissingErrorMessage(t *testing.T) {
	_, chain := setupErrorsTest(t)

	req := solo.NewCallParams(errors.Contract.Name, errors.FuncRegisterError.Name).
		WithGasBudget(100_000)

	_, _, err := chain.PostRequestSyncTx(req, nil)

	testError := &vmerrors.Error{}
	require.ErrorAs(t, err, &testError)

	typedError := err.(*vmerrors.Error)
	require.Equal(t, typedError.Definition(), *commonerrors.ErrUntypedError)

	require.Equal(t, err.Error(), "cannot decode key 'm': cannot decode nil bytes")
}

func TestSuccessfulRegisterError(t *testing.T) {
	_, chain := setupErrorsTest(t)

	req := solo.NewCallParams(errors.Contract.Name, errors.FuncRegisterError.Name, errors.ParamErrorMessageFormat, "poof").
		WithGasBudget(100_000)

	_, _, err := chain.PostRequestSyncTx(req, nil)

	require.NoError(t, err)
}

func TestRetrievalOfErrorMessage(t *testing.T) {
	errorMessageToTest := "poof"

	_, chain := setupErrorsTest(t)

	req := solo.NewCallParams(errors.Contract.Name, errors.FuncRegisterError.Name, errors.ParamErrorMessageFormat, errorMessageToTest).
		WithGasBudget(100_000)

	_, dict, err := chain.PostRequestSyncTx(req, nil)

	require.NoError(t, err)

	kv := kvdecoder.New(dict)
	errorId := kv.MustGetUint16(errors.ParamErrorId)
	contract := kv.MustGetUint32(errors.ParamContractHname)

	req = solo.NewCallParams(errors.Contract.Name, errors.FuncGetErrorMessageFormat.Name, errors.ParamContractHname, contract, errors.ParamErrorId, errorId).
		WithGasBudget(100_000)

	_, dict, err = chain.PostRequestSyncTx(req, nil)

	require.NoError(t, err)

	message := dict.MustGet(errors.ParamErrorMessageFormat)

	require.Equal(t, string(message), errorMessageToTest)
}
