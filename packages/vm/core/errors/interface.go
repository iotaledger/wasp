package errors

import (
	"github.com/iotaledger/wasp/v2/packages/isc"
	"github.com/iotaledger/wasp/v2/packages/isc/coreutil"
)

var Contract = coreutil.NewContract(coreutil.CoreContractErrors)

var (
	FuncRegisterError = coreutil.NewEP11(Contract, "registerError",
		coreutil.Field[string]("errorMessageFormat"),
		coreutil.Field[isc.VMErrorCode]("vmErrorCode"),
	)

	ViewGetErrorMessageFormat = coreutil.NewViewEP11(Contract, "getErrorMessageFormat",
		coreutil.Field[isc.VMErrorCode]("vmErrorCode"),
		coreutil.Field[string]("errorMessageFormat"),
	)
)

const (
	prefixErrorTemplateMap = "a" // covered in: TestSuccessfulRegisterError
)
