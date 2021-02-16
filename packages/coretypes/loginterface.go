package coretypes

type LogInterface interface {
	Infof(format string, param ...interface{})
	Debugf(format string, param ...interface{})
	Panicf(format string, param ...interface{})
}
