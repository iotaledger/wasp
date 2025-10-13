package main

import (
	"bytes"
	"errors"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/tools/wasp-cli/authentication"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/chain"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/cli/config"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/cli/setup"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/codec"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/metrics"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/peering"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/wallet"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/waspcmd"
)

// TestHarness provides a simple testing environment for cobra commands
type TestHarness struct {
	t       *testing.T
	rootCmd *cobra.Command
	stdout  *bytes.Buffer
	stderr  *bytes.Buffer
}

// NewTestHarness creates a new test harness
func NewTestHarness(t *testing.T) *TestHarness {
	h := &TestHarness{
		t:      t,
		stdout: &bytes.Buffer{},
		stderr: &bytes.Buffer{},
	}

	h.setupRootCommand()
	return h
}

// setupRootCommand creates a root command for testing
func (h *TestHarness) setupRootCommand() {
	h.rootCmd = &cobra.Command{
		Use:           "wasp-cli",
		Short:         "wasp-cli test harness",
		Long:          "wasp-cli test harness for unit testing",
		Version:       "test-version",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Skip config reading in tests
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	// mimic main.go persistent flags
	h.rootCmd.PersistentFlags().Bool("skip-version-check", true, "skip-version-check")
	h.rootCmd.PersistentFlags().BoolVar(&config.PrettyPrintConfig, "format-config", true, "format the config file when saving")
	h.rootCmd.PersistentFlags().Uint32P("address-index", "i", 0, "address index")

	// Set up output capture
	h.rootCmd.SetOut(h.stdout)
	h.rootCmd.SetErr(h.stderr)

	// Add subcommands
	h.addSubcommands()
}

// addSubcommands adds CLI subcommands
func (h *TestHarness) addSubcommands() {
	// Initialize log first
	log.Init(h.rootCmd)

	// Add subcommands
	setup.Init(h.rootCmd)
	authentication.Init(h.rootCmd)
	waspcmd.Init(h.rootCmd)
	wallet.Init(h.rootCmd)
	chain.Init(h.rootCmd)
	codec.Init(h.rootCmd)
	peering.Init(h.rootCmd)
	metrics.Init(h.rootCmd)
}

// RunCommand executes a command with the given arguments
func (h *TestHarness) RunCommand(args ...string) (string, error) {
	// Reset buffers
	h.stdout.Reset()
	h.stderr.Reset()

	// Set command arguments
	h.rootCmd.SetArgs(args)

	// Execute command
	err := h.rootCmd.Execute()

	// Get output
	stdout := h.stdout.String()
	stderr := h.stderr.String()

	// Combine stdout and stderr
	output := stdout
	if stderr != "" {
		if output != "" {
			output += "\n"
		}
		output += stderr
	}

	return output, err
}

// MustRunCommand executes a command and requires it to succeed
func (h *TestHarness) MustRunCommand(args ...string) string {
	output, err := h.RunCommand(args...)
	require.NoError(h.t, err, "Command failed: %v", args)
	return output
}

// RunCommandExpectError executes a command expecting it to fail
func (h *TestHarness) RunCommandExpectError(args ...string) (string, error) {
	output, err := h.RunCommand(args...)
	require.Error(h.t, err, "Command should have failed: %v", args)
	return output, err
}

// AssertOutputContains checks that output contains expected string
func (h *TestHarness) AssertOutputContains(output, expected string) {
	require.Contains(h.t, output, expected, "Output should contain: %s", expected)
}

// SetupTestConfig sets up minimal test configuration
func (h *TestHarness) SetupTestConfig() {
	config.PrettyPrintConfig = false
}

// CleanupTestConfig cleans up test configuration
func (h *TestHarness) CleanupTestConfig() {
	// Reset any test state
}

// TestCommandFramework demonstrates the command testing framework
func TestCommandFramework(t *testing.T) {
	h := NewTestHarness(t)
	h.SetupTestConfig()
	defer h.CleanupTestConfig()

	t.Run("root help command", func(t *testing.T) {
		output := h.MustRunCommand("--help")
		h.AssertOutputContains(output, "wasp-cli")
		h.AssertOutputContains(output, "Usage:")
	})

	t.Run("wallet help command", func(t *testing.T) {
		output := h.MustRunCommand("wallet", "--help")
		h.AssertOutputContains(output, "wallet")
		h.AssertOutputContains(output, "address")
	})

	t.Run("chain help command", func(t *testing.T) {
		output := h.MustRunCommand("chain", "--help")
		h.AssertOutputContains(output, "chain")
		h.AssertOutputContains(output, "deploy")
	})

	t.Run("invalid command", func(t *testing.T) {
		_, err := h.RunCommandExpectError("invalid-command")
		require.Error(t, err)
	})

	t.Run("auth help command", func(t *testing.T) {
		output := h.MustRunCommand("auth", "--help")
		h.AssertOutputContains(output, "auth")
		h.AssertOutputContains(output, "login")
	})
}

// TestCommandValidation tests command parameter validation
func TestCommandValidation(t *testing.T) {
	h := NewTestHarness(t)
	h.SetupTestConfig()
	defer h.CleanupTestConfig()

	t.Run("chain deploy missing required flag", func(t *testing.T) {
		_, err := h.RunCommandExpectError("chain", "deploy")
		require.Error(t, err)
		// Should fail because --chain flag is required
	})

	t.Run("codec encode shows help when missing subcommand", func(t *testing.T) {
		output := h.MustRunCommand("codec", "encode")
		require.Contains(t, output, "Usage:")
	})
}

// TestOutputFormats tests different output formats
func TestOutputFormats(t *testing.T) {
	h := NewTestHarness(t)
	h.SetupTestConfig()
	defer h.CleanupTestConfig()

	t.Run("help output format", func(t *testing.T) {
		output := h.MustRunCommand("--help")

		// Should contain usage information
		h.AssertOutputContains(output, "Usage:")
		h.AssertOutputContains(output, "Available Commands:")
		h.AssertOutputContains(output, "Flags:")
	})

	t.Run("subcommand help format", func(t *testing.T) {
		output := h.MustRunCommand("wallet", "--help")

		// Should contain wallet-specific help
		h.AssertOutputContains(output, "wallet")
		h.AssertOutputContains(output, "Usage:")
	})
}

// TestErrorHandling tests error scenarios
func TestErrorHandling(t *testing.T) {
	h := NewTestHarness(t)
	h.SetupTestConfig()
	defer h.CleanupTestConfig()

	t.Run("unknown command", func(t *testing.T) {
		_, err := h.RunCommandExpectError("unknown-command")
		require.Error(t, err)
		require.Contains(t, err.Error(), "unknown command")
	})

	t.Run("unknown subcommand", func(t *testing.T) {
		_, err := h.RunCommandExpectError("wallet", "unknown-subcommand")
		require.Error(t, err)
	})
}

// MockExample demonstrates basic mocking concept
type MockService struct {
	Value  string
	Called bool
	Error  error
}

func (m *MockService) GetValue() (string, error) {
	m.Called = true
	if m.Error != nil {
		return "", m.Error
	}
	return m.Value, nil
}

func TestMockingConcept(t *testing.T) {
	mock := &MockService{Value: "test-value"}

	// Test successful call
	value, err := mock.GetValue()
	require.NoError(t, err)
	require.Equal(t, "test-value", value)
	require.True(t, mock.Called)

	// Test error case
	mock.Error = errors.New("test error")
	mock.Called = false

	_, err = mock.GetValue()
	require.Error(t, err)
	require.True(t, mock.Called)
}
