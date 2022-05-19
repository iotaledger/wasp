package coreerrors

import (
	"github.com/iotaledger/wasp/packages/iscp"
)

var coreErrorCollection = NewCoreErrorCollection()

func Register(messageFormat string) *iscp.VMErrorTemplate {
	template, err := coreErrorCollection.Register(messageFormat)
	if err != nil {
		panic(err)
	}
	return template
}

func All() ErrorCollection {
	return coreErrorCollection
}
