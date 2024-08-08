package errors

import (
	"github.com/iotaledger/wasp/packages/isc/coreutil"
	"github.com/iotaledger/wasp/packages/kv/codec"
)

var Contract = coreutil.NewContract(coreutil.CoreContractErrors)

var (
	FuncRegisterError = coreutil.NewEP11(Contract, "registerError",
		coreutil.FieldWithCodec(codec.String),
		coreutil.FieldWithCodec(codec.VMErrorCode),
	)

	ViewGetErrorMessageFormat = coreutil.NewViewEP11(Contract, "getErrorMessageFormat",
		coreutil.FieldWithCodec(codec.VMErrorCode),
		coreutil.FieldWithCodec(codec.String),
	)
)

const (
	prefixErrorTemplateMap = "a" // covered in: TestSuccessfulRegisterError
)
