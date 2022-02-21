package coreerrors

import (
	"github.com/iotaledger/wasp/packages/iscp"
)

type ErrorCollection interface {
	Get(errorId uint16) (*iscp.VMErrorTemplate, error)
	Register(messageFormat string) (*iscp.VMErrorTemplate, error)
}
