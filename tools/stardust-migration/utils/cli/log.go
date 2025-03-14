package cli

import "log"

func Logf(format string, s ...interface{}) {
	print(func() {
		log.Printf(format, s...)
	})
}

func Logln(v ...interface{}) {
	print(func() {
		log.Println(v...)
	})
}

func Log(v ...interface{}) {
	print(func() {
		log.Print(v...)
	})
}

var DebugLoggingEnabled = false

func DebugLogf(format string, s ...interface{}) {
	if DebugLoggingEnabled {
		Logf(format, s...)
	}
}

func DebugLogln(v ...interface{}) {
	if DebugLoggingEnabled {
		Logln(v...)
	}
}

func DebugLog(v ...interface{}) {
	if DebugLoggingEnabled {
		Log(v...)
	}
}
