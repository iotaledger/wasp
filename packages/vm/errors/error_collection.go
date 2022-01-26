package errors

import "golang.org/x/xerrors"

type ErrorCollection interface {
	Get(errorId int) *ErrorDefinition
	Create(errorId int, params ...interface{}) (*BlockError, error)
}

type DefaultErrorCollection struct {
	Errors map[int]ErrorDefinition
}

func (e *DefaultErrorCollection) Get(errorId int) *ErrorDefinition {
	if errorDefinition, ok := e.Errors[errorId]; ok {
		return &errorDefinition
	}

	return nil
}

func (e *DefaultErrorCollection) Create(errorId int, params ...interface{}) (*BlockError, error) {
	if errorDefinition := e.Get(errorId); errorDefinition == nil {
		return nil, xerrors.Errorf("No error found with id %v", errorId)
	} else {
		errorInstance := errorDefinition.Create(params)
		return &errorInstance, nil
	}
}

var GeneralErrorCollection = DefaultErrorCollection{
	Errors: map[int]ErrorDefinition{
		1: NewBlockErrorDefinition(1, "Failed to load %v"),
		2: NewBlockErrorDefinition(2, "Could not read from %v"),
	},
}
