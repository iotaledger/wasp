package errors

import (
	"golang.org/x/xerrors"
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

func (e *ErrorCollection) Create(errorId uint16, params ...interface{}) (*BlockError, error) {
	if errorDefinition, err := e.Get(errorId); err != nil {
		return nil, err
	} else if errorDefinition == nil {
		return nil, xerrors.Errorf("No error found with id %v", errorId)
	} else {
		errorInstance := errorDefinition.Create(params...)
		return &errorInstance, nil
	}
}

func (e *ErrorCollection) Register(errorId uint16, messageFormat string) (*ErrorDefinition, error) {
	if _, exists := e.errors[errorId]; exists {
		return nil, xerrors.Errorf("Error with id %v already registered", errorId)
	}

	errorDefinition := ErrorDefinition{
		id:            errorId,
		messageFormat: messageFormat,
	}

	e.errors[errorId] = errorDefinition

	return &errorDefinition, nil
}
