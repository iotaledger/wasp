// Package testerrors contains helpers for contract error testing
package testerrors

import (
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
)

var (
	ContractName = "testerrors"
	Contract     = coreutil.NewContract(ContractName)
)

var (
	FuncRegisterErrors        = Contract.Func("register_errors")
	FuncThrowErrorWithoutArgs = Contract.Func("throw_error_without_args")
	FuncThrowErrorWithArgs    = Contract.Func("throw_error_with_args")
	FuncThrowUntypedError     = Contract.Func("throw_untyped_error")
)

var (
	MessageToTest = "Test error message %d"
	Error         *isc.VMErrorTemplate
)

var Processor = Contract.Processor(nil,
	FuncRegisterErrors.WithHandler(func(ctx isc.Sandbox) isc.CallArguments {
		Error = ctx.RegisterError(MessageToTest)

		return nil
	}),
	FuncThrowErrorWithoutArgs.WithHandler(func(ctx isc.Sandbox) isc.CallArguments {
		panic(Error.Create())
	}),
	FuncThrowErrorWithArgs.WithHandler(func(ctx isc.Sandbox) isc.CallArguments {
		panic(Error.Create(uint8(42)))
	}),
	FuncThrowUntypedError.WithHandler(func(ctx isc.Sandbox) isc.CallArguments {
		panic("untyped error")
	}),
)
