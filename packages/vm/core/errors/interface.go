package errors

import (
	"github.com/iotaledger/wasp/packages/iscp/coreutil"
)

var Contract = coreutil.NewContract(coreutil.CoreContractErrors, "Errors contract")

//nolint:deadcode,varcheck // prefixBlockRegistry and prefixControlAddresses are not used, can those be removed?
const (
	prefixBlockRegistry = string('a' + iota)
	prefixControlAddresses
	prefixErrorTemplateMap
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
