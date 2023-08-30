package errors

import (
	"github.com/iotaledger/wasp/packages/isc/coreutil"
)

var Contract = coreutil.NewContract(coreutil.CoreContractErrors)

var (
	FuncRegisterError = coreutil.Func("registerError")

	ViewGetErrorMessageFormat = coreutil.ViewFunc("getErrorMessageFormat")
)

// request parameters
const (
	ParamErrorCode          = "c"
	ParamErrorMessageFormat = "m"
)

const (
	prefixErrorTemplateMap = "a"
)
