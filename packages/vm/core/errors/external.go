package errors

import (
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

// SandboxErrorMessageResolver has the signature of VMErrorMessageResolver to provide a way to resolve the error format
func SandboxErrorMessageResolver(ctx iscp.SandboxView) func(*iscp.VMError) (string, error) {
	return func(errorToResolve *iscp.VMError) (string, error) {
		params := dict.New()
		params.Set(ParamContractHname, codec.EncodeHname(errorToResolve.PrefixId()))
		params.Set(ParamErrorId, codec.EncodeUint16(errorToResolve.Id()))

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
