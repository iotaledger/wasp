package errors

import (
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/vm/core/errors/coreerrors"
)

var Processor = Contract.Processor(nil,
	FuncRegisterError.WithHandler(funcRegisterError),
	ViewGetErrorMessageFormat.WithHandler(funcGetErrorMessageFormat),
)

func SetInitialState(state kv.KVStore) {
	// does not do anything
}

func funcRegisterError(ctx isc.Sandbox, errorMessageFormat string) dict.Dict {
	ctx.Log().Debugf("Registering error")

	if errorMessageFormat == "" {
		panic(coreerrors.ErrMessageFormatEmpty)
	}

	e := NewStateErrorCollectionWriter(ctx.State(), ctx.Contract())
	template, err := e.Register(errorMessageFormat)
	ctx.RequireNoError(err)

	return dict.Dict{ParamErrorCode: codec.VMErrorCode.Encode(template.Code())}
}

func funcGetErrorMessageFormat(ctx isc.SandboxView, code isc.VMErrorCode) string {
	template, ok := getErrorMessageFormat(ctx.StateR(), code)
	if !ok {
		panic(coreerrors.ErrErrorNotFound)
	}

	return template.MessageFormat()
}

func getErrorMessageFormat(state kv.KVStoreReader, code isc.VMErrorCode) (*isc.VMErrorTemplate, bool) {
	var e coreerrors.ErrorCollection
	if code.ContractID == isc.VMCoreErrorContractID {
		e = coreerrors.All()
	} else {
		e = NewStateErrorCollectionReader(state, code.ContractID)
	}
	return e.Get(code.ID)
}
