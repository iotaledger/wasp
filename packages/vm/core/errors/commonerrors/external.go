package commonerrors

import errors "github.com/iotaledger/wasp/packages/vm/vmerrors"

var globalErrorCollection IErrorCollection = NewErrorCollection()

func RegisterGlobalError(messageFormat string) *errors.ErrorDefinition {
	errorId, err := errors.GetErrorIdFromMessageFormat(messageFormat)

	if err != nil {
		panic(err)
	}

	errorDefinition, err := globalErrorCollection.Register(errorId, messageFormat)

	if err != nil {
		panic(err)
	}

	return errorDefinition
}

func GetGlobalErrorCollection() IErrorCollection {
	return globalErrorCollection
}

// SandboxErrorMessageResolver has the signature of ErrorMessageResolver to provide a way to resolve the error format
