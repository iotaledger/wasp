package testcore

import (
	"strings"
	"testing"

	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/solo"
	"github.com/iotaledger/wasp/packages/vm"
	"github.com/iotaledger/wasp/packages/vm/core"
	"github.com/iotaledger/wasp/packages/vm/core/errors"
	"github.com/iotaledger/wasp/packages/vm/core/errors/coreerrors"
	"github.com/stretchr/testify/require"
)

var errorMessageToTest = "Test error message %v"

var (
	errorContractName = "ErrorContract"
	errorContract     = coreutil.NewContract(errorContractName, "error contract")
)

var (
	funcRegisterErrors        = coreutil.Func("register_errors")
	funcThrowErrorWithoutArgs = coreutil.Func("throw_error_without_args")
	funcThrowErrorWithArgs    = coreutil.Func("throw_error_with_args")
)

var testError *iscp.VMErrorTemplate

var errorContractProcessor = errorContract.Processor(nil,
	funcRegisterErrors.WithHandler(func(ctx iscp.Sandbox) dict.Dict {
		testError = ctx.RegisterError(errorMessageToTest)

		return nil
	}),
	funcThrowErrorWithoutArgs.WithHandler(func(ctx iscp.Sandbox) dict.Dict {
		panic(testError.Create())
	}),
	funcThrowErrorWithArgs.WithHandler(func(ctx iscp.Sandbox) dict.Dict {
		panic(testError.Create(42.0))
	}),
)

func setupErrorsTest(t *testing.T) (*solo.Solo, *solo.Chain) {
	core.PrintWellKnownHnames()
	env := solo.New(t, &solo.InitOptions{AutoAdjustDustDeposit: true, Debug: true}).WithNativeContract(errorContractProcessor)
	chain, _, _ := env.NewChainExt(nil, 100_000, "chain1")
	err := chain.DeployContract(nil, errorContract.Name, errorContract.ProgramHash)

	require.NoError(t, err)

	chain.MustDepositIotasToL2(10_000_000, nil)
	defer chain.Log().Sync()

	chain.CheckChain()

	return env, chain
}

func setupErrorsTestWithoutFunds(t *testing.T) (*solo.Solo, *solo.Chain) {
	core.PrintWellKnownHnames()
	env := solo.New(t, &solo.InitOptions{AutoAdjustDustDeposit: true, Debug: true})
	chain, _, _ := env.NewChainExt(nil, 1, "chain1")

	chain.MustDepositIotasToL2(1, nil)
	defer chain.Log().Sync()
	chain.CheckChain()

	return env, chain
}

// Panics and returned errors will eventually land into the error handling hook.
// Typical xerror/error types will be wrapped into an UnresolvedVMError type (Err ErrUntypedError)
// Panicked vmerrors will be stored as is.
// The first test validates a typed vmerror UnresolvedVMError (Not enough Gas)
// The second test validates the wrapped generic ErrUntypedError
func TestErrorWithCustomError(t *testing.T) {
	_, chain := setupErrorsTestWithoutFunds(t)

	req := solo.NewCallParams(errors.Contract.Name, errors.FuncRegisterError.Name).
		WithGasBudget(1)

	_, _, err := chain.PostRequestSyncTx(req, nil)

	testError := &iscp.VMError{}
	require.ErrorAs(t, err, &testError)

	typedError := err.(*iscp.VMError)
	require.Equal(t, typedError.AsTemplate(), vm.ErrGasBudgetDetail)
}

// This test does not supply the required kv pair 'ParamErrorMessageFormat' which makes the kvdecoder fail with an xerror
func TestPanicDueMissingErrorMessage(t *testing.T) {
	_, chain := setupErrorsTest(t)

	req := solo.NewCallParams(errors.Contract.Name, errors.FuncRegisterError.Name).
		WithGasBudget(100_000)

	_, _, err := chain.PostRequestSyncTx(req, nil)

	testError := &iscp.VMError{}
	require.ErrorAs(t, err, &testError)

	typedError := err.(*iscp.VMError)
	require.Equal(t, typedError.AsTemplate(), coreerrors.ErrUntypedError)

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
	_, chain := setupErrorsTest(t)

	req := solo.NewCallParams(errors.Contract.Name, errors.FuncRegisterError.Name, errors.ParamErrorMessageFormat, errorMessageToTest).
		WithGasBudget(100_000)

	_, dict, err := chain.PostRequestSyncTx(req, nil)
	require.NoError(t, err)

	errorCode := codec.MustDecodeVMErrorCode(dict.MustGet(errors.ParamErrorCode))

	req = solo.NewCallParams(errors.Contract.Name, errors.FuncGetErrorMessageFormat.Name,
		errors.ParamErrorCode, errorCode,
	).
		WithGasBudget(100_000)

	_, dict, err = chain.PostRequestSyncTx(req, nil)
	require.NoError(t, err)

	message := dict.MustGet(errors.ParamErrorMessageFormat)

	require.Equal(t, string(message), errorMessageToTest)
}

func TestErrorRegistrationWithCustomContract(t *testing.T) {
	_, chain := setupErrorsTest(t)

	req := solo.NewCallParams(errorContract.Name, funcRegisterErrors.Name).
		WithGasBudget(100_000)

	_, _, err := chain.PostRequestSyncTx(req, nil)

	require.NoError(t, err)

	require.Equal(t, testError.Code().ID, iscp.GetErrorIDFromMessageFormat(errorMessageToTest))
}

func TestPanicWithCustomContractWithArgs(t *testing.T) {
	_, chain := setupErrorsTest(t)

	// Register error
	req := solo.NewCallParams(errorContract.Name, funcRegisterErrors.Name).
		WithGasBudget(100_000)

	_, _, err := chain.PostRequestSyncTx(req, nil)

	require.NoError(t, err)

	// Throw error
	req = solo.NewCallParams(errorContract.Name, funcThrowErrorWithArgs.Name).
		WithGasBudget(100_000)

	_, _, err = chain.PostRequestSyncTx(req, nil)

	errorTestType := &iscp.VMError{}
	require.ErrorAs(t, err, &errorTestType)

	typedError := err.(*iscp.VMError)

	require.Error(t, err)
	require.Equal(t, testError.Code().ID, iscp.GetErrorIDFromMessageFormat(errorMessageToTest))
	require.Equal(t, testError.Code().ContractID, typedError.Code().ContractID)

	// Further, this error will add the arg '42'
	require.True(t, strings.HasSuffix(err.Error(), "42"))
}

func TestPanicWithCustomContractWithoutArgs(t *testing.T) {
	_, chain := setupErrorsTest(t)

	// Register error
	req := solo.NewCallParams(errorContract.Name, funcRegisterErrors.Name).
		WithGasBudget(100_000)

	_, _, err := chain.PostRequestSyncTx(req, nil)

	require.NoError(t, err)

	// Throw error
	req = solo.NewCallParams(errorContract.Name, funcThrowErrorWithoutArgs.Name).
		WithGasBudget(100_000)

	_, _, err = chain.PostRequestSyncTx(req, nil)

	errorTestType := &iscp.VMError{}
	require.ErrorAs(t, err, &errorTestType)

	typedError := err.(*iscp.VMError)

	require.Error(t, err)
	require.Equal(t, testError.Code().ID, iscp.GetErrorIDFromMessageFormat(errorMessageToTest))
	require.Equal(t, testError.Code().ContractID, typedError.Code().ContractID)

	t.Log(err.Error())

	// This error throws without an expected arg. Therefore, the output will end with '%!v(MISSING)'
	require.True(t, strings.HasSuffix(err.Error(), "%!v(MISSING)"))
}

func TestUnresolvedErrorIsStoredInReceiptAndIsEqualToVMErrorWithoutArgs(t *testing.T) {
	_, chain := setupErrorsTest(t)

	// Register error
	req := solo.NewCallParams(errorContract.Name, funcRegisterErrors.Name).
		WithGasBudget(100_000)

	_, _, err := chain.PostRequestSyncTx(req, nil)

	require.NoError(t, err)

	// Throw error
	req = solo.NewCallParams(errorContract.Name, funcThrowErrorWithoutArgs.Name).
		WithGasBudget(100_000)

	_, _, err = chain.PostRequestSyncTx(req, nil)

	receipt := chain.LastReceipt()
	typedError := err.(*iscp.VMError)

	errorTestType := &iscp.VMError{}
	receiptErrorTestType := &iscp.UnresolvedVMError{}

	require.Error(t, receipt.Error.AsGoError())

	require.ErrorAs(t, err, &errorTestType)
	require.ErrorAs(t, receipt.Error, &receiptErrorTestType)

	require.EqualValues(t, receipt.Error.Code(), typedError.Code())
	require.EqualValues(t, receipt.Error.Hash(), typedError.Hash())
	require.EqualValues(t, receipt.Error.Params(), typedError.Params())
}

func TestUnresolvedErrorIsStoredInReceiptAndIsEqualToVMErrorWithArgs(t *testing.T) {
	_, chain := setupErrorsTest(t)

	// Register error
	req := solo.NewCallParams(errorContract.Name, funcRegisterErrors.Name).
		WithGasBudget(100_000)

	_, _, err := chain.PostRequestSyncTx(req, nil)

	require.NoError(t, err)

	// Throw error
	req = solo.NewCallParams(errorContract.Name, funcThrowErrorWithArgs.Name).
		WithGasBudget(100_000)

	_, _, err = chain.PostRequestSyncTx(req, nil)

	receipt := chain.LastReceipt()
	typedError := err.(*iscp.VMError)

	errorTestType := &iscp.VMError{}
	receiptErrorTestType := &iscp.UnresolvedVMError{}

	require.Error(t, receipt.Error.AsGoError())

	require.ErrorAs(t, err, &errorTestType)
	require.ErrorAs(t, receipt.Error, &receiptErrorTestType)

	require.EqualValues(t, receipt.Error.Code(), typedError.Code())
	require.EqualValues(t, receipt.Error.Hash(), typedError.Hash())
	require.Equal(t, receipt.Error.Params(), typedError.Params())
}

func TestIsComparer(t *testing.T) {
	template := iscp.NewVMErrorTemplate(iscp.NewVMErrorCode(1234, 1), "fooBar")
	vmerror := template.Create()
	vmerrorUnresolved := vmerror.AsUnresolvedError()

	template2 := iscp.NewVMErrorTemplate(iscp.NewVMErrorCode(4321, 2), "barFoo")

	require.True(t, iscp.VMErrorIs(template, vmerrorUnresolved))
	require.True(t, iscp.VMErrorIs(template, vmerror))
	require.True(t, iscp.VMErrorIs(vmerror, template))
	require.True(t, iscp.VMErrorIs(vmerror, vmerrorUnresolved))
	require.True(t, iscp.VMErrorIs(vmerrorUnresolved, template))
	require.True(t, iscp.VMErrorIs(vmerrorUnresolved, vmerror))

	require.False(t, iscp.VMErrorIs(template2, vmerrorUnresolved))
	require.False(t, iscp.VMErrorIs(template2, vmerror))
	require.False(t, iscp.VMErrorIs(template2, template))
	require.False(t, iscp.VMErrorIs(vmerror, template2))
	require.False(t, iscp.VMErrorIs(vmerrorUnresolved, template2))
}
