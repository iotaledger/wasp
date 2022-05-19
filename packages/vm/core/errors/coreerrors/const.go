package coreerrors

var (
	ErrUntypedError           = Register("%v")
	ErrErrorAlreadyRegistered = Register("Error with id %v already registered")
	ErrErrorMessageTooLong    = Register("Error message is too long").Create()
	ErrMessageFormatEmpty     = Register("Error message is empty").Create()
)
