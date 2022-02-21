package errors

import (
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
)

var Contract = coreutil.NewContract(coreutil.CoreContractErrors, "Errors contract")

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
