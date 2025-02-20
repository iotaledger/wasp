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

func DebugLogf(format string, s ...interface{}) {
}

func DebugLogln(v ...interface{}) {
}

func DebugLog(v ...interface{}) {
}
