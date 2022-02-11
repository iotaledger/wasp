package errors

import (
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
	error2 "github.com/iotaledger/wasp/packages/vm/vmerrors"
)

func SandboxErrorMessageResolver(ctx iscp.SandboxView) func(*error2.Error) (string, error) {
	return func(errorToResolve *error2.Error) (string, error) {
		params := dict.New()
		params.Set(ParamContractHname, codec.EncodeUint32(errorToResolve.PrefixId))
		params.Set(ParamErrorId, codec.EncodeUint16(errorToResolve.Id))

		ret := ctx.Call(Contract.Hname(), FuncGetErrorMessageFormat.Hname(), params)

		errorMessageFormat, err := ret.Get(ParamErrorMessageFormat)

		if err != nil {
			return "", err
		}

		errorMessageFormatString, err := codec.DecodeString(errorMessageFormat)

		if err != nil {
			return "", err
		}

		return errorMessageFormatString, nil
	}
}
