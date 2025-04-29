//go:build linux

// Package util provides general utility functions and structures,
// including operating system specific utilities for process management.
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
