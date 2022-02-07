package errors

import (
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
	"math"
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
	e := NewStateErrorCollectionWriter(ctx.State(), ctx.Caller().Hname())

	params := kvdecoder.New(ctx.Params())
	errorId := params.MustGetUint16(ParamErrorId)
	errorMessageFormat := params.MustGetString(ParamErrorMessageFormat)

	errorDefinition, err := e.Register(errorId, errorMessageFormat)

	if err != nil {
		panic(err)
	}

	ret := dict.New()
	ret.Set(ParamErrorId, codec.EncodeUint16(errorId))
	ret.Set(ParamErrorDefinitionAdded, codec.EncodeBool(errorDefinition != nil))

	return ret
}

func funcGetErrorMessageFormat(ctx iscp.SandboxView) dict.Dict {
	params := kvdecoder.New(ctx.Params())

	contract := params.MustGetHname(ParamContractHname)
	errorId := params.MustGetUint16(ParamErrorId)

	var e IErrorCollection

	if contract == math.MaxUint32 {
		e = globalErrorCollection
	} else {
		e = NewStateErrorCollectionReader(ctx.State(), contract)
	}

	errorDefinition, err := e.Get(errorId)

	if err != nil {
		panic(err)
	}

	ret := dict.New()
	ret.Set(ParamErrorMessageFormat, codec.EncodeString(errorDefinition.MessageFormat()))

	return ret
}
