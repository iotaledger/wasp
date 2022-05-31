//go:build linux

package util

import (
	"os/exec"
	"syscall"
)

func TerminateCmdWhenTestStops(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Pdeathsig: syscall.SIGTERM,
	}
}
