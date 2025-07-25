// Package coreerrors provides a centralized way to manage and register error templates
// for the IOTA Smart Contract VM. It defines an error collection mechanism that enables
// standardized error handling across the VM core, allowing for consistent error reporting
// and management. The package exposes functions to register new error templates and
// retrieve all registered errors, making it easier to maintain and reference VM errors
// throughout the codebase.
package coreerrors

import (
	"github.com/iotaledger/wasp/v2/packages/isc"
)

var coreErrorCollection = NewCoreErrorCollection()

func Register(messageFormat string) *isc.VMErrorTemplate {
	template, err := coreErrorCollection.Register(messageFormat)
	if err != nil {
		panic(err)
	}
	return template
}

func All() ErrorCollection {
	return coreErrorCollection
}
