package errors

import (
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/vm/core/errors/coreerrors"
)

var Processor = Contract.Processor(nil,
	FuncRegisterError.WithHandler(funcRegisterError),
	ViewGetErrorMessageFormat.WithHandler(funcGetErrorMessageFormat),
)

func (s *StateWriter) SetInitialState() {
	// does not do anything
}

func funcRegisterError(ctx isc.Sandbox, errorMessageFormat string) isc.VMErrorCode {
	ctx.Log().Debugf("Registering error")

	if errorMessageFormat == "" {
		panic(coreerrors.ErrMessageFormatEmpty)
	}

	e := NewStateWriterFromSandbox(ctx).ErrorCollection(ctx.Contract())
	template, err := e.Register(errorMessageFormat)
	ctx.RequireNoError(err)

	return template.Code()
}

func funcGetErrorMessageFormat(ctx isc.SandboxView, code isc.VMErrorCode) string {
	template, ok := NewStateReaderFromSandbox(ctx).getErrorMessageFormat(code)
	if !ok {
		panic(coreerrors.ErrErrorNotFound)
	}
	return template.MessageFormat()
}

func (s *StateReader) getErrorMessageFormat(code isc.VMErrorCode) (*isc.VMErrorTemplate, bool) {
	var e coreerrors.ErrorCollection
	if code.ContractID == isc.VMCoreErrorContractID {
		e = coreerrors.All()
	} else {
		e = s.ErrorCollection(code.ContractID)
	}
	return e.Get(code.ID)
}
