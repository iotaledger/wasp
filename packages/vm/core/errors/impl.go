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

	errorDefinition, err := e.Register(errorMessageFormat)

	ctx.RequireNoError(err)

	ret := dict.New()
	ret.Set(ParamContractHname, codec.EncodeHname(ctx.Caller().Hname()))
	ret.Set(ParamErrorId, codec.EncodeUint16(errorDefinition.Id()))
	ret.Set(ParamErrorDefinitionAdded, codec.EncodeBool(errorDefinition != nil))

	return ret
}

func funcGetErrorMessageFormat(ctx iscp.SandboxView) dict.Dict {
	params := kvdecoder.New(ctx.Params())

	contract := params.MustGetUint32(ParamContractHname)
	errorId := params.MustGetUint16(ParamErrorId)

	var e coreerrors.ErrorCollection

	// All errors with a MaxUint32 contract id will be considered as core errors.
	if contract == iscp.VMCoreErrorID {
		e = coreerrors.All()
	} else {
		e = NewStateErrorCollectionReader(ctx.State(), iscp.Hname(contract))
	}

	errorDefinition, err := e.Get(errorId)

	ctx.RequireNoError(err)

	ret := dict.New()
	ret.Set(ParamErrorMessageFormat, codec.EncodeString(errorDefinition.MessageFormat()))

	return ret
}
