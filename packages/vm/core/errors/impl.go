package errors

import (
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
	"github.com/iotaledger/wasp/packages/vm/core/errors/coreerrors"
)

var Processor = Contract.Processor(initialize,
	FuncRegisterError.WithHandler(funcRegisterError),
	FuncGetErrorMessageFormat.WithHandler(funcGetErrorMessageFormat),
)

func initialize(ctx iscp.Sandbox) dict.Dict {
	ctx.Log().Debugf("errors.initialize.success hname = %s", Contract.Hname().String())
	return nil
}

func funcRegisterError(ctx iscp.Sandbox) dict.Dict {
	ctx.Log().Debugf("Registering error")
	e := NewStateErrorCollectionWriter(ctx.State(), ctx.Caller().Hname())

	params := kvdecoder.New(ctx.Params())
	errorMessageFormat := params.MustGetString(ParamErrorMessageFormat)

	if len(errorMessageFormat) == 0 {
		panic(coreerrors.ErrMessageFormatEmpty)
	}

	template, err := e.Register(errorMessageFormat)
	ctx.RequireNoError(err)

	return dict.Dict{ParamErrorCode: codec.EncodeVMErrorCode(template.Code())}
}

func funcGetErrorMessageFormat(ctx iscp.SandboxView) dict.Dict {
	code := codec.MustDecodeVMErrorCode(ctx.Params().MustGet(ParamErrorCode))

	var e coreerrors.ErrorCollection

	if code.ContractID == iscp.VMCoreErrorContractID {
		e = coreerrors.All()
	} else {
		e = NewStateErrorCollectionReader(ctx.State(), code.ContractID)
	}

	template, err := e.Get(code.ID)
	ctx.RequireNoError(err)

	return dict.Dict{ParamErrorMessageFormat: codec.EncodeString(template.MessageFormat())}
}
