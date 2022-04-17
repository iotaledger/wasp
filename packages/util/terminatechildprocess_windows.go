//go:build windows

package util

import (
	"os/exec"
)

func TerminateCmdWhenTestStops(cmd *exec.Cmd) {
	// do nothing, SysprocAttr is not available on windows
	// maybe there is a way to achieve a similar result, but as of now
	// just be aware that child processes might be left hanging if
	// the test process is forcibly stopped (i.e. it times out)
}
