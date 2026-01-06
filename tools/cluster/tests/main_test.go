package tests

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// TestMain builds the wasp-cli binary once per package and shares the path via waspCliBinPath.
func TestMain(m *testing.M) {
	// Allow bypassing the build if CI provides a prebuilt binary
	if p := os.Getenv("WASP_CLI_BIN"); p != "" {
		waspCliBinPath = p
	} else {
		// Build to a temp file once for this package run
		tmpDir, err := os.MkdirTemp(os.TempDir(), "wasp-cli-bin-*")
		if err != nil {
			panic(err)
		}
		// Determine repo root based on this source file location
		_, thisFile, _, _ := runtime.Caller(0)
		repoRoot := filepath.Clean(filepath.Join(filepath.Dir(thisFile), "../../.."))
		binPath := filepath.Join(tmpDir, "wasp-cli")

		// Resolve the 'go' binary from PATH and derive GOROOT via 'go env' to avoid deprecated runtime.GOROOT()
		goBin, err := exec.LookPath("go")
		if err != nil {
			panic("failed to locate 'go' in PATH: " + err.Error())
		}

		// Query the active toolchain's GOROOT via 'go env'
		gorootCmd := exec.CommandContext(context.Background(), goBin, "env", "GOROOT")
		gorootOut := new(bytes.Buffer)
		gorootCmd.Stdout = gorootOut
		gorootCmd.Stderr = gorootOut
		if err := gorootCmd.Run(); err != nil {
			panic("failed to resolve GOROOT via 'go env': " + gorootOut.String())
		}
		goroot := strings.TrimSpace(gorootOut.String())

		buildCmd := exec.CommandContext(context.Background(), goBin, "build", "-o", binPath, "./tools/wasp-cli")
		buildCmd.Dir = repoRoot
		// Sanitize environment to align with this toolchain: set GOROOT and clear GOTOOLDIR if present
		env := os.Environ()
		// override GOROOT with the value reported by 'go env'
		env = append(env, "GOROOT="+goroot)
		// clear GOTOOLDIR to let the tool find the right tools under GOROOT/bin
		// (cannot remove easily; set to empty to neutralize)
		env = append(env, "GOTOOLDIR=")
		buildCmd.Env = env
		out := new(bytes.Buffer)
		buildCmd.Stdout = out
		buildCmd.Stderr = out
		if err := buildCmd.Run(); err != nil {
			panic("failed to build wasp-cli: (cd " + buildCmd.Dir + " && " + strings.Join(buildCmd.Args, " ") + ")\n" + out.String())
		}
		waspCliBinPath = binPath
	}

	os.Exit(m.Run())
}
