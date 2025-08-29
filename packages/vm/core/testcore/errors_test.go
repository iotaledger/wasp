// excluded temporarily because of compilation errors

package testcore

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/solo"
	"github.com/iotaledger/wasp/v2/packages/testutil/testdbhash"
	"github.com/iotaledger/wasp/v2/packages/vm/core/corecontracts"
	"github.com/iotaledger/wasp/v2/packages/vm/core/errors"
	"github.com/iotaledger/wasp/v2/packages/vm/core/errors/coreerrors"
	"github.com/iotaledger/wasp/v2/packages/vm/core/evm"
	"github.com/iotaledger/wasp/v2/packages/vm/core/governance"
	"github.com/iotaledger/wasp/v2/packages/vm/core/testcore/contracts/testerrors"
)

func setupErrorsTest(t *testing.T) *solo.Chain {
	corecontracts.PrintWellKnownHnames()
	env := solo.New(t, &solo.InitOptions{Debug: true})
	chain, _ := env.NewChainExt(nil, 0, "chain1", evm.DefaultChainID, governance.DefaultBlockKeepAmount, nil, nil)

	chain.MustDepositBaseTokensToL2(10*isc.Million, nil)
	defer chain.Log().Shutdown()

	chain.CheckChain()

	return chain
}

// Panics and returned errors will eventually land into the error handling hook.
// Typical xerror/error types will be wrapped into an UnresolvedVMError type (Err ErrUntypedError)
// Panicked vmerrors will be stored as is.

func TestUntypedError(t *testing.T) {
	chain := setupErrorsTest(t)

	_, _, _, _, err := chain.PostRequestSyncTx(
		solo.NewCallParams(testerrors.FuncThrowUntypedError.Message(nil)),
		nil,
	)

	testError := &isc.VMError{}
	require.ErrorAs(t, err, &testError)
	require.ErrorContains(t, err, "untyped error")
}

// This test does not supply the required kv pair 'ParamErrorMessageFormat' which makes the kvdecoder fail with an xerror
func TestPanicDueMissingErrorMessage(t *testing.T) {
	chain := setupErrorsTest(t)

	req := solo.NewCallParamsEx(errors.Contract.Name, errors.FuncRegisterError.Name).
		WithGasBudget(100_000)

	_, _, _, _, err := chain.PostRequestSyncTx(req, nil)

	testError := &isc.VMError{}
	require.ErrorAs(t, err, &testError)

	typedError := err.(*isc.VMError)
	require.Equal(t, typedError.AsTemplate(), coreerrors.ErrUntypedError)

	require.ErrorContains(t, err, "index out of range")
}

func TestSuccessfulRegisterError(t *testing.T) {
	chain := setupErrorsTest(t)

	req := solo.NewCallParams(errors.FuncRegisterError.Message("poof")).
		WithGasBudget(100_000)

	_, _, _, _, err := chain.PostRequestSyncTx(req, nil)
	require.NoError(t, err)

	testdbhash.VerifyContractStateHash(chain.Env, errors.Contract, "", t.Name())
}

func TestRetrievalOfErrorMessage(t *testing.T) {
	chain := setupErrorsTest(t)

	errorCode, err := errors.FuncRegisterError.Call(testerrors.MessageToTest, func(msg isc.Message) (isc.CallArguments, error) {
		req := solo.NewCallParams(msg).
			WithGasBudget(100_000)

		d, innerErr := chain.PostRequestSync(req, nil)
		return d, innerErr
	})
	require.NoError(t, err)

	message, err := errors.ViewGetErrorMessageFormat.Call(errorCode, func(msg isc.Message) (isc.CallArguments, error) {
		req := solo.NewCallParams(msg).
			WithGasBudget(100_000)

		d, innerErr := chain.PostRequestSync(req, nil)
		return d, innerErr
	})
	require.NoError(t, err)
	require.Equal(t, testerrors.MessageToTest, message)
}

func TestErrorRegistrationWithCustomContract(t *testing.T) {
	chain := setupErrorsTest(t)

	req := solo.NewCallParams(testerrors.FuncRegisterErrors.Message(nil)).
		WithGasBudget(100_000)

	_, err := chain.PostRequestSync(req, nil)

	require.NoError(t, err)

	require.Equal(t, testerrors.Error.Code().ID, isc.GetErrorIDFromMessageFormat(testerrors.MessageToTest))
}

func TestPanicWithCustomContractWithArgs(t *testing.T) {
	chain := setupErrorsTest(t)

	// Register error
	req := solo.NewCallParams(testerrors.FuncRegisterErrors.Message(nil)).
		WithGasBudget(100_000)

	_, err := chain.PostRequestSync(req, nil)
	require.NoError(t, err)

	// Throw error
	req = solo.NewCallParams(testerrors.FuncThrowErrorWithArgs.Message(nil)).
		WithGasBudget(100_000)

	_, err = chain.PostRequestSync(req, nil)
	require.Error(t, err)
	require.True(t, strings.HasSuffix(err.Error(), "42"))

	errorTestType := &isc.VMError{}
	require.ErrorAs(t, err, &errorTestType)

	typedError := err.(*isc.VMError)

	require.Equal(t, testerrors.Error.Code().ID, isc.GetErrorIDFromMessageFormat(testerrors.MessageToTest))
	require.Equal(t, testerrors.Error.Code().ContractID, typedError.Code().ContractID)
}

func TestPanicWithCustomContractWithoutArgs(t *testing.T) {
	chain := setupErrorsTest(t)

	// Register error
	req := solo.NewCallParams(testerrors.FuncRegisterErrors.Message(nil)).
		WithGasBudget(100_000)

	_, err := chain.PostRequestSync(req, nil)

	require.NoError(t, err)

	// Throw error
	req = solo.NewCallParams(testerrors.FuncThrowErrorWithoutArgs.Message(nil)).
		WithGasBudget(100_000)

	_, err = chain.PostRequestSync(req, nil)

	errorTestType := &isc.VMError{}
	require.ErrorAs(t, err, &errorTestType)

	typedError := err.(*isc.VMError)

	require.Error(t, err)
	require.Equal(t, testerrors.Error.Code().ID, isc.GetErrorIDFromMessageFormat(testerrors.MessageToTest))
	require.Equal(t, testerrors.Error.Code().ContractID, typedError.Code().ContractID)

	t.Log(err.Error())

	// This error throws without an expected arg. Therefore, the output will end with '%!v(MISSING)'
	require.True(t, strings.HasSuffix(err.Error(), "%!d(MISSING)"))
}

func TestUnresolvedErrorIsStoredInReceiptAndIsEqualToVMErrorWithoutArgs(t *testing.T) {
	chain := setupErrorsTest(t)

	// Register error
	req := solo.NewCallParams(testerrors.FuncRegisterErrors.Message(nil)).
		WithGasBudget(100_000)

	_, err := chain.PostRequestSync(req, nil)

	require.NoError(t, err)

	// Throw error
	req = solo.NewCallParams(testerrors.FuncThrowErrorWithoutArgs.Message(nil)).
		WithGasBudget(100_000)

	_, err = chain.PostRequestSync(req, nil)

	receipt := chain.LastReceipt()
	typedError := err.(*isc.VMError)

	errorTestType := &isc.VMError{}
	receiptErrorTestType := &isc.UnresolvedVMError{}

	require.Error(t, receipt.Error.AsGoError())

	require.ErrorAs(t, err, &errorTestType)
	require.ErrorAs(t, receipt.Error, &receiptErrorTestType)

	require.EqualValues(t, receipt.Error.Code(), typedError.Code())
	require.Empty(t, typedError.Params())
	require.Empty(t, receipt.Error.Params)
}

func TestUnresolvedErrorIsStoredInReceiptAndIsEqualToVMErrorWithArgs(t *testing.T) {
	chain := setupErrorsTest(t)

	// Register error
	req := solo.NewCallParams(testerrors.FuncRegisterErrors.Message(nil)).
		WithGasBudget(100_000)

	_, err := chain.PostRequestSync(req, nil)

	require.NoError(t, err)

	// Throw error
	req = solo.NewCallParams(testerrors.FuncThrowErrorWithArgs.Message(nil)).
		WithGasBudget(100_000)

	_, err = chain.PostRequestSync(req, nil)

	receipt := chain.LastReceipt()
	typedError := err.(*isc.VMError)

	errorTestType := &isc.VMError{}
	receiptErrorTestType := &isc.UnresolvedVMError{}

	require.Error(t, receipt.Error.AsGoError())

	require.ErrorAs(t, err, &errorTestType)
	require.ErrorAs(t, receipt.Error, &receiptErrorTestType)

	require.EqualValues(t, receipt.Error.Code(), typedError.Code())
	require.Equal(t, receipt.Error.Params, typedError.Params())
}

func TestIsComparer(t *testing.T) {
	template := isc.NewVMErrorTemplate(isc.NewVMErrorCode(1234, 1), "fooBar")
	vmerror := template.Create()
	vmerrorUnresolved := vmerror.AsUnresolvedError()

	template2 := isc.NewVMErrorTemplate(isc.NewVMErrorCode(4321, 2), "barFoo")

	require.True(t, isc.VMErrorIs(template, vmerrorUnresolved))
	require.True(t, isc.VMErrorIs(template, vmerror))
	require.True(t, isc.VMErrorIs(vmerror, template))
	require.True(t, isc.VMErrorIs(vmerror, vmerrorUnresolved))
	require.True(t, isc.VMErrorIs(vmerrorUnresolved, template))
	require.True(t, isc.VMErrorIs(vmerrorUnresolved, vmerror))

	require.False(t, isc.VMErrorIs(template2, vmerrorUnresolved))
	require.False(t, isc.VMErrorIs(template2, vmerror))
	require.False(t, isc.VMErrorIs(template2, template))
	require.False(t, isc.VMErrorIs(vmerror, template2))
	require.False(t, isc.VMErrorIs(vmerrorUnresolved, template2))
}
