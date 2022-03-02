package errors

import (
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
)

var Contract = coreutil.NewContract(coreutil.CoreContractErrors, "Errors contract")

const (
	prefixBlockRegistry = string('a' + iota)
	prefixControlAddresses
	prefixErrorTemplateMap
)

var (
	FuncRegisterError         = coreutil.Func("registerError")
	FuncGetErrorMessageFormat = coreutil.ViewFunc("getErrorMessageFormat")
)

// parameters
const (
	ParamErrorCode          = "c"
	ParamErrorMessageFormat = "m"
)
