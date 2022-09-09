package coreerrors

import (
	"github.com/iotaledger/wasp/packages/isc"
)

// CoreErrorCollection implements ErrorCollection. Is used for global errors. Does not require vm context state.
type CoreErrorCollection map[uint16]*isc.VMErrorTemplate

func NewCoreErrorCollection() ErrorCollection {
	return CoreErrorCollection{}
}

func (e CoreErrorCollection) Get(errorID uint16) (*isc.VMErrorTemplate, error) {
	if template, ok := e[errorID]; ok {
		return template, nil
	}
	return nil, nil
}

func (e CoreErrorCollection) Register(messageFormat string) (*isc.VMErrorTemplate, error) {
	if len(messageFormat) > isc.VMErrorMessageLimit {
		return nil, ErrErrorMessageTooLong
	}

	errorID := isc.GetErrorIDFromMessageFormat(messageFormat)

	if _, exists := e[errorID]; exists {
		return nil, ErrErrorAlreadyRegistered.Create(errorID)
	}

	e[errorID] = isc.NewVMErrorTemplate(isc.NewCoreVMErrorCode(errorID), messageFormat)

	return e[errorID], nil
}
