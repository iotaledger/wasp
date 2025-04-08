package cli

import (
	"fmt"
	"sync"
)

var statusBar string
var m sync.Mutex

func UpdateStatusBarf(format string, s ...interface{}) {
	UpdateStatusBar(fmt.Sprintf(format, s...))
}

func UpdateStatusBar(s string) {
	m.Lock()
	defer m.Unlock()

	clearStatusBar()
	fmt.Print(s)
	statusBar = s
}

func ClearStatusBar() {
	UpdateStatusBar("")
}

func clearStatusBar() {
	fmt.Print("\033[2K\r")
}

func print(printFunc func()) {
	m.Lock()
	defer m.Unlock()

	fmt.Print("\033[2K\r")
	printFunc()
	fmt.Print(statusBar)
}

func Println(v ...interface{}) {
	print(func() {
		fmt.Print(append(v, "\n")...)
	})
}

func Printf(format string, s ...interface{}) {
	if format[len(format)-1] == '\n' {
		Println(fmt.Sprintf(format[:len(format)-1], s...))
		return
	}

	Println(fmt.Sprintf(format, s...))
}
