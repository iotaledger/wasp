package errors

import (
	"github.com/iotaledger/wasp/packages/isc"
	"github.com/iotaledger/wasp/packages/isc/coreutil"
)

var Contract = coreutil.NewContract(coreutil.CoreContractErrors)

var (
	FuncRegisterError = coreutil.NewEP11(Contract, "registerError",
		coreutil.Field[string](),
		coreutil.Field[isc.VMErrorCode](),
	)

	ViewGetErrorMessageFormat = coreutil.NewViewEP11(Contract, "getErrorMessageFormat",
		coreutil.Field[isc.VMErrorCode](),
		coreutil.Field[string](),
	)
)

const (
	prefixErrorTemplateMap = "a" // covered in: TestSuccessfulRegisterError
)
