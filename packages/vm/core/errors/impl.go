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

func initialize(ctx iscp.Sandbox) (dict.Dict, error) {
	ctx.Log().Debugf("errors.initialize.success hname = %s", Contract.Hname().String())
	return nil, nil
}

func funcRegisterError(ctx iscp.Sandbox) (dict.Dict, error) {
	params := kvdecoder.New(ctx.Params())
	errorId := params.MustGetUint16(ParamErrorId)
	errorMessageFormat := params.MustGetString(ParamErrorMessageFormat)

	success, err := AddErrorDefinition(ctx.State(), ctx.Caller().Hname(), errorId, errorMessageFormat)

	if err != nil {
		return nil, err
	}

	ret := dict.New()
	ret.Set(ParamErrorDefinitionAdded, codec.EncodeBool(success))

	return ret, nil
}
