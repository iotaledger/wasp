package errors

import (
	"github.com/iotaledger/wasp/packages/iscp"
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
)

var Contract = coreutil.NewContract(coreutil.CoreContractError, "Error contract")

const (
	prefixBlockRegistry = string('a' + iota)
	prefixControlAddresses
)

var (
	FuncRegisterError         = coreutil.Func("registerError")
	FuncGetErrorMessageFormat = coreutil.ViewFunc("getErrorMessageFormat")
)

const (
	// parameters
	ParamErrorDefinitionMap   = "e"
	ParamErrorId              = "i"
	ParamContractHname        = "h"
	ParamErrorMessageFormat   = "m"
	ParamErrorDefinitionAdded = "a"
)

type IErrorCollection interface {
	Get(errorId uint16) (*iscp.ErrorDefinition, error)
	Register(errorId uint16, messageFormat string) (*iscp.ErrorDefinition, error)
}
