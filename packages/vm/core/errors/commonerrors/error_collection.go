package commonerrors

import (
	"github.com/iotaledger/wasp/packages/vm/vmerrors"
	"math"
)

// ErrorCollection implements IErrorCollection. Is used for global errors. Does not require vm context state.
type ErrorCollection struct {
	errors map[uint16]*vmerrors.ErrorDefinition
}

func NewErrorCollection() IErrorCollection {
	return &ErrorCollection{
		errors: map[uint16]*vmerrors.ErrorDefinition{},
	}
}

func (e *ErrorCollection) Get(errorId uint16) (*vmerrors.ErrorDefinition, error) {
	if errorDefinition, ok := e.errors[errorId]; ok {
		return errorDefinition, nil
	}

	return nil, nil
}

func (e *ErrorCollection) Register(errorId uint16, messageFormat string) (*vmerrors.ErrorDefinition, error) {
	if len(messageFormat) > math.MaxUint16 {
		return nil, ErrErrorMessageTooLong
	}

	if _, exists := e.errors[errorId]; exists {
		return nil, ErrErrorAlreadyRegistered.Create(errorId)
	}

	e.errors[errorId] = vmerrors.NewErrorDefinition(math.MaxUint32, errorId, messageFormat)

	return e.errors[errorId], nil
}
