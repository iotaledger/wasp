package coreerrors

import (
	"github.com/iotaledger/wasp/packages/iscp"
	"math"
)

// CoreErrorCollection implements IErrorCollection. Is used for global errors. Does not require vm context state.
type CoreErrorCollection map[uint16]*iscp.VMErrorTemplate

func NewCoreErrorCollection() ErrorCollection {
	return &CoreErrorCollection{}
}

func (e *CoreErrorCollection) Get(errorId uint16) (*iscp.VMErrorTemplate, error) {

	if errorDefinition, ok := (*e)[errorId]; ok {
		return errorDefinition, nil
	}

	return nil, nil
}

func (e *CoreErrorCollection) Register(messageFormat string) (*iscp.VMErrorTemplate, error) {
	if len(messageFormat) > iscp.VMErrorMessageLimit {
		return nil, ErrErrorMessageTooLong
	}

	errorId := iscp.GetErrorIdFromMessageFormat(messageFormat)

	if _, exists := (*e)[errorId]; exists {
		return nil, ErrErrorAlreadyRegistered.Create(errorId)
	}

	(*e)[errorId] = iscp.NewVMErrorTemplate(math.MaxUint32, errorId, messageFormat)

	return (*e)[errorId], nil
}
