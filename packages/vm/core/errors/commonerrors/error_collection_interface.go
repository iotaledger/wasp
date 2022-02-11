package commonerrors

import (
	"github.com/iotaledger/wasp/packages/vm/vmerrors"
)

type IErrorCollection interface {
	Get(errorId uint16) (*vmerrors.ErrorDefinition, error)
	Register(errorId uint16, messageFormat string) (*vmerrors.ErrorDefinition, error)
}
