package coreerrors

var (
	ErrUntypedError          = Register("%s")
	ErrErrorTemplateConflict = Register("error with id %d conflicts with an already registered error")
	ErrErrorMessageTooLong   = Register("error message is too long").Create()
	ErrErrorNotFound         = Register("error not found").Create()
	ErrMessageFormatEmpty    = Register("error message is empty").Create()
)
