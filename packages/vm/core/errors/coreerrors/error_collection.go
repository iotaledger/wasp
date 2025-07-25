package coreerrors

import (
	"github.com/iotaledger/wasp/v2/packages/isc"
)

// CoreErrorCollection implements ErrorCollection. Is used for global errors. Does not require vm context state.
type CoreErrorCollection map[uint16]*isc.VMErrorTemplate

func NewCoreErrorCollection() ErrorCollectionWriter {
	return CoreErrorCollection{}
}

func (e CoreErrorCollection) Get(errorID uint16) (*isc.VMErrorTemplate, bool) {
	template, ok := e[errorID]
	return template, ok
}

func (e CoreErrorCollection) Register(messageFormat string) (*isc.VMErrorTemplate, error) {
	if len(messageFormat) > isc.VMErrorMessageLimit {
		return nil, ErrErrorMessageTooLong
	}

	errorID := isc.GetErrorIDFromMessageFormat(messageFormat)

	if t, exists := e[errorID]; exists && t.MessageFormat() != messageFormat {
		return nil, ErrErrorTemplateConflict.Create(errorID)
	}

	e[errorID] = isc.NewVMErrorTemplate(isc.NewCoreVMErrorCode(errorID), messageFormat)

	return e[errorID], nil
}
