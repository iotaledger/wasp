package webapi_validation

type CoreAccountsValidation struct {
	ValidationContext
}

func NewCoreAccountsValidation(validationContext ValidationContext) CoreAccountsValidation {
	return CoreAccountsValidation{
		ValidationContext: validationContext,
	}
}

func (c *CoreAccountsValidation) Validate(stateIndex uint32) {
}
