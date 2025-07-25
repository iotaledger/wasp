package errors

import (
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/vm/core/errors/coreerrors"
)

// ViewCaller is a generic interface for any function that can call views
type ViewCaller func(msg isc.Message) (isc.CallArguments, error)

func GetMessageFormat(code isc.VMErrorCode, callView ViewCaller) (string, error) {
	ret, err := callView(ViewGetErrorMessageFormat.Message(code))
	if err != nil {
		return "", err
	}

	errorMessage, err := ViewGetErrorMessageFormat.DecodeOutput(ret)
	if err != nil {
		return "", err
	}

	return errorMessage, nil
}

func Resolve(e *isc.UnresolvedVMError, callView ViewCaller) (*isc.VMError, error) {
	if e == nil {
		return nil, nil
	}

	messageFormat, err := GetMessageFormat(e.Code(), callView)
	if err != nil {
		return nil, err
	}

	return isc.NewVMErrorTemplate(e.Code(), messageFormat).Create(e.Params...), nil
}

func (s *StateReader) Resolve(e *isc.UnresolvedVMError) (*isc.VMError, error) {
	if e == nil {
		return nil, nil
	}
	template, ok := s.getErrorMessageFormat(e.Code())
	if !ok {
		return nil, coreerrors.ErrErrorNotFound
	}
	return isc.NewVMErrorTemplate(e.Code(), template.MessageFormat()).Create(e.Params...), nil
}
