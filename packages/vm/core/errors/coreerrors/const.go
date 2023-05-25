package coreerrors

var (
	ErrUntypedError          = Register("%v")
	ErrErrorTemplateConflict = Register("Error with id %d conflicts with an already registered error")
	ErrErrorMessageTooLong   = Register("Error message is too long").Create()
	ErrErrorNotFound         = Register("Error not found").Create()
	ErrMessageFormatEmpty    = Register("Error message is empty").Create()
)
