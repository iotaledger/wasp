package coreerrors

import (
	"github.com/iotaledger/wasp/packages/isc"
)

type ErrorCollection interface {
	Get(errorID uint16) (*isc.VMErrorTemplate, error)
	Register(messageFormat string) (*isc.VMErrorTemplate, error)
}
