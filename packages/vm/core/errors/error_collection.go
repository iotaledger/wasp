package errors

import (
	"golang.org/x/xerrors"
	"math"
)

// ErrorCollection implements IErrorCollection. Is used for global errors. Does not require vm context state.
type ErrorCollection struct {
	errors map[uint16]ErrorDefinition
}

func NewErrorCollection() IErrorCollection {
	return &ErrorCollection{
		errors: map[uint16]ErrorDefinition{},
	}
}

func (e *ErrorCollection) Get(errorId uint16) (*ErrorDefinition, error) {
	if errorDefinition, ok := e.errors[errorId]; ok {
		return &errorDefinition, nil
	}

	return nil, nil
}

func (e *ErrorCollection) Register(errorId uint16, messageFormat string) (*ErrorDefinition, error) {
	if _, exists := e.errors[errorId]; exists {
		return nil, xerrors.Errorf("Error with id %v already registered", errorId)
	}

	errorDefinition := ErrorDefinition{
		prefixId:      math.MaxUint32, // Global errors should be a constant MaxUint32 instead of being a misleading zero.
		id:            errorId,
		messageFormat: messageFormat,
	}

	e.errors[errorId] = errorDefinition

	return &errorDefinition, nil
}
