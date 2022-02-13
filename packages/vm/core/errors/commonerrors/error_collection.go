package commonerrors

import (
	errors "github.com/iotaledger/wasp/packages/vm/vmerrors"
	"math"
)

// ErrorCollection implements IErrorCollection. Is used for global errors. Does not require vm context state.
type ErrorCollection struct {
	errors map[uint16]*errors.ErrorDefinition
}

func NewErrorCollection() IErrorCollection {
	return &ErrorCollection{
		errors: map[uint16]*errors.ErrorDefinition{},
	}
}

func (e *ErrorCollection) Get(errorId uint16) (*errors.ErrorDefinition, error) {
	if errorDefinition, ok := e.errors[errorId]; ok {
		return errorDefinition, nil
	}

	return nil, nil
}

func (e *ErrorCollection) Register(errorId uint16, messageFormat string) (*errors.ErrorDefinition, error) {
	if len(messageFormat) > math.MaxUint16 {
		return nil, ErrErrorMessageTooLong
	}

	if _, exists := e.errors[errorId]; exists {
		return nil, ErrErrorAlreadyRegistered.CreateTyped(errorId)
	}

	e.errors[errorId] = errors.NewErrorDefinition(math.MaxUint32, errorId, messageFormat)

	return e.errors[errorId], nil
}
