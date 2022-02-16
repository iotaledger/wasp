package coreerrors

import (
	"github.com/iotaledger/wasp/packages/iscp"
)

var coreErrorCollection ErrorCollection = NewCoreErrorCollection()

func Register(messageFormat string) *iscp.VMErrorTemplate {
	errorDefinition, err := coreErrorCollection.Register(messageFormat)

	if err != nil {
		panic(err)
	}

	return errorDefinition
}

func All() ErrorCollection {
	return coreErrorCollection
}
