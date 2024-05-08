package errors

import (
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/kv/codec"
)

var Contract = coreutil.NewContract(coreutil.CoreContractErrors)

var (
	FuncRegisterError = coreutil.NewEP1(Contract, "registerError",
		coreutil.FieldWithCodec(ParamErrorMessageFormat, codec.String),
	)

	ViewGetErrorMessageFormat = coreutil.NewViewEP11(Contract, "getErrorMessageFormat",
		coreutil.FieldWithCodec(ParamErrorCode, codec.VMErrorCode),
		coreutil.FieldWithCodec(ParamErrorMessageFormat, codec.String),
	)
)

// request parameters
const (
	ParamErrorCode          = "c"
	ParamErrorMessageFormat = "m"
)

const (
	prefixErrorTemplateMap = "a" // covered in: TestSuccessfulRegisterError
)
