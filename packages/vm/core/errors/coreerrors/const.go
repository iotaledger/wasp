package coreerrors

var ErrUntypedError = Register("%v")
var ErrErrorAlreadyRegistered = Register("Error with id %v already registered")
var ErrErrorMessageTooLong = Register("Error message is too long").Create()
var ErrMessageFormatEmpty = Register("Error message is empty").Create()
