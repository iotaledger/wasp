package commonerrors

var ErrUntypedError = RegisterGlobalError("%v")
var ErrErrorAlreadyRegistered = RegisterGlobalError("Error with id %v already registered")
var ErrErrorMessageTooLong = RegisterGlobalError("Error message is too long").Create()
var ErrMessageFormatEmpty = RegisterGlobalError("Error message is empty").Create()
