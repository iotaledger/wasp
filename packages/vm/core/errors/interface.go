package errors

import (
	"github.com/iotaledger/wasp/packages/isc/coreutil"
)

var Contract = coreutil.NewContract(coreutil.CoreContractErrors, "Errors contract")

const (
	prefixErrorTemplateMap = "a"
)

var (
	FuncRegisterError         = coreutil.Func("registerError")
	ViewGetErrorMessageFormat = coreutil.ViewFunc("getErrorMessageFormat")
)

// parameters
const (
	ParamErrorCode          = "c"
	ParamErrorMessageFormat = "m"
)
