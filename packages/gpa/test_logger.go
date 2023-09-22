package gpa

import "fmt"

// Useful in tests, to make some warnings apparent.
type panicLogger struct{}

func NewPanicLogger() Logger {
	return &panicLogger{}
}

func (*panicLogger) Warnf(msg string, args ...any) {
	panic(fmt.Errorf(msg, args...))
}
