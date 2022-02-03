package errors

import (
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	"github.com/iotaledger/wasp/packages/kv/kvdecoder"
)

var Processor = Contract.Processor(initialize,
	FuncRegisterError.WithHandler(funcRegisterError),
)

func initialize(ctx iscp.Sandbox) dict.Dict {
	ctx.Log().Debugf("errors.initialize.success hname = %s", Contract.Hname().String())
	return nil
}

func funcRegisterError(ctx iscp.Sandbox) dict.Dict {
	e := NewStateErrorCollection(ctx.State(), ctx.Caller().Hname())

	params := kvdecoder.New(ctx.Params())
	errorId := params.MustGetUint16(ParamErrorId)
	errorMessageFormat := params.MustGetString(ParamErrorMessageFormat)

	errorDefinition, err := e.Register(errorId, errorMessageFormat)

	if err != nil {
		return nil
	}

	ret := dict.New()
	ret.Set(ParamErrorDefinitionAdded, codec.EncodeBool(errorDefinition != nil))

	return ret
}
