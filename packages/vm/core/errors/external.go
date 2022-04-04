package errors

import (
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/kv/codec"
	"github.com/iotaledger/wasp/packages/kv/dict"
)

// ViewCaller is a generic interface for any function that can call views
type ViewCaller func(contractName string, funcName string, params dict.Dict) (dict.Dict, error)

func GetMessageFormat(code iscp.VMErrorCode, callView ViewCaller) (string, error) {
	ret, err := callView(Contract.Name, FuncGetErrorMessageFormat.Name, dict.Dict{
		ParamErrorCode: codec.EncodeVMErrorCode(code),
	})
	if err != nil {
		return "", err
	}
	return codec.DecodeString(ret.MustGet(ParamErrorMessageFormat))
}

func ResolveErrorCode(e *iscp.VMErrorCode, callView ViewCaller) (*iscp.VMErrorTemplate, error) {
	if e == nil {
		return nil, nil
	}

	messageFormat, err := GetMessageFormat(*e, callView)

	if err != nil {
		return nil, err
	}

	return iscp.NewVMErrorTemplate(*e, messageFormat), nil
}

func Resolve(e *iscp.UnresolvedVMError, callView ViewCaller) (*iscp.VMError, error) {
	if e == nil {
		return nil, nil
	}

	messageFormat, err := GetMessageFormat(e.Code(), callView)
	if err != nil {
		return nil, err
	}

	return iscp.NewVMErrorTemplate(e.Code(), messageFormat).Create(e.Params()...), nil
}
