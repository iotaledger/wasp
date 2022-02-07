package errors

import (
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

const GlobalErrorStart uint16 = 4096

var globalErrorCollection IErrorCollection = NewErrorCollection()

func RegisterGlobalError(errorId uint16, messageFormat string) *iscp.ErrorDefinition {
	errorDefinition, err := globalErrorCollection.Register(errorId, messageFormat)

	if err != nil {
		panic(err)
	}

	return errorDefinition
}

func GetGlobalError(errorId uint16) (*iscp.ErrorDefinition, error) {
	return globalErrorCollection.Get(errorId)
}

func GetStateError(partition kv.KVStore, errorId uint16, contract iscp.Hname) (*iscp.ErrorDefinition, error) {
	e := NewStateErrorCollectionReader(partition, contract)

	errorDefinition, err := e.Get(errorId)

	if err != nil {
		return nil, err
	}

	return errorDefinition, nil
}

func Panic(definition iscp.ErrorDefinition, params ...interface{}) {
	panic(definition.Create(params...))
}

// SandboxErrorMessageResolver has the signature of ErrorMessageResolver to provide a way to resolve the error format
func SandboxErrorMessageResolver(ctx iscp.SandboxView) func(*iscp.Error) (string, error) {
	return func(errorToResolve *iscp.Error) (string, error) {
		params := dict.New()
		params.Set(ParamContractHname, codec.EncodeHname(errorToResolve.PrefixId))
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
