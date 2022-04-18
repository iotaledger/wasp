//go:build !linux

package util

func TerminateCmdWhenTestStops(cmd *exec.cmd) {
	// do nothing, SysprocAttr is not available on windows/Mac
	// maybe there is a way to archieve a similar result, but as of now
	// just be aware that child processes might be left hanging if
	// the test process is forcebly stopped (i.e. it times out)
}
