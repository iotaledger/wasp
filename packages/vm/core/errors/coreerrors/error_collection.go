package coreerrors

import (
	"github.com/iotaledger/wasp/packages/iscp"
)

// CoreErrorCollection implements ErrorCollection. Is used for global errors. Does not require vm context state.
type CoreErrorCollection map[uint16]*iscp.VMErrorTemplate

func NewCoreErrorCollection() ErrorCollection {
	return CoreErrorCollection{}
}

func (e CoreErrorCollection) Get(errorID uint16) (*iscp.VMErrorTemplate, error) {
	if template, ok := e[errorID]; ok {
		return template, nil
	}
	return nil, nil
}

func (e CoreErrorCollection) Register(messageFormat string) (*iscp.VMErrorTemplate, error) {
	if len(messageFormat) > iscp.VMErrorMessageLimit {
		return nil, ErrErrorMessageTooLong
	}

	errorID := iscp.GetErrorIDFromMessageFormat(messageFormat)

	if _, exists := e[errorID]; exists {
		return nil, ErrErrorAlreadyRegistered.Create(errorID)
	}

	e[errorID] = iscp.NewVMErrorTemplate(iscp.NewCoreVMErrorCode(errorID), messageFormat)

	return e[errorID], nil
}
