package test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"github.com/iotaledger/wasp/v2/tools/wasp-cli/authentication"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/chain"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/cli/config"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/cli/setup"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/codec"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/disrec"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/inspection"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/log"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/metrics"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/peering"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/wallet"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/waspcmd"
)

// sliceContains checks if a slice contains a string
func sliceContains(slice []string, val string) bool {
	for _, a := range slice {
		if a == val {
			return true
		}
	}
	return false
}

// TestHarness provides a testing environment for cobra commands
type TestHarness struct {
	t       *testing.T
	rootCmd *cobra.Command
	stdout  *bytes.Buffer
	stderr  *bytes.Buffer

	// Mocks
	MockL1Client     *MockL1Client
	MockWallet       *MockWallet
	MockChainService *MockChainService
}

// NewTestHarness creates a new test harness with mocked dependencies
func NewTestHarness(t *testing.T) *TestHarness {
	h := &TestHarness{
		t:      t,
		stdout: &bytes.Buffer{},
		stderr: &bytes.Buffer{},

		// Initialize mocks
		MockL1Client:     NewMockL1Client(),
		MockWallet:       NewMockWallet(),
		MockChainService: NewMockChainService(),
	}

	h.setupRootCommand()
	return h
}

// setupRootCommand creates a root command similar to main.go but with test configuration
func (h *TestHarness) setupRootCommand() {
	h.rootCmd = &cobra.Command{
		Use:           "wasp-cli",
		Short:         "wasp-cli test harness",
		Long:          "wasp-cli test harness for unit testing",
		Version:       "test-version",
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Skip config reading and validation in tests
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	// mimic main.go persistent flags for parity
	// These are safe no-ops for tests but make flags available
	h.rootCmd.PersistentFlags().Bool("skip-version-check", true, "skip-version-check")
	h.rootCmd.PersistentFlags().BoolVar(&config.PrettyPrintConfig, "format-config", true, "format the config file when saving")
	h.rootCmd.PersistentFlags().Uint32P("address-index", "i", 0, "address index")

	// Set up output capture
	h.rootCmd.SetOut(h.stdout)
	h.rootCmd.SetErr(h.stderr)

	// Add all subcommands
	h.addSubcommands()
}

// addSubcommands adds all the CLI subcommands to the root command
func (h *TestHarness) addSubcommands() {
	// Initialize log first
	log.Init(h.rootCmd)

	// Add all subcommands like in main.go
	setup.Init(h.rootCmd)
	authentication.Init(h.rootCmd)
	waspcmd.Init(h.rootCmd)
	wallet.Init(h.rootCmd)
	chain.Init(h.rootCmd)
	codec.Init(h.rootCmd)
	peering.Init(h.rootCmd)
	metrics.Init(h.rootCmd)
	disrec.Init(h.rootCmd)
	inspection.Init(h.rootCmd)
}

// RunCommand executes a command with the given arguments and returns output
func (h *TestHarness) RunCommand(args ...string) (string, error) {
	// Reset buffers
	h.stdout.Reset()
	h.stderr.Reset()

	// Set command arguments
	h.rootCmd.SetArgs(args)

	// Execute command
	err := h.rootCmd.Execute()

	// Collect output streams
	stdout := strings.TrimSpace(h.stdout.String())
	stderr := strings.TrimSpace(h.stderr.String())

	// Combine stdout and stderr for output
	output := strings.TrimSpace(strings.Join([]string{stdout, stderr}, "\n"))
	output = strings.Trim(output, "\n ")

	// If JSON flags are present, format output similar to production behavior
	argsStr := strings.Join(args, " ")
	jsonPretty := strings.Contains(argsStr, " --json") || strings.HasSuffix(argsStr, "--json") || sliceContains(args, "--json")
	jsonCompact := strings.Contains(argsStr, " --json-compact") || strings.HasSuffix(argsStr, "--json-compact") || sliceContains(args, "--json-compact")

	if jsonPretty || jsonCompact {
		// If command already produced JSON, pass it through unchanged
		if h.IsValidJSON(stdout) && stderr == "" {
			return stdout, err
		}
		// Otherwise wrap stdout/stderr and error into a JSON object
		payload := map[string]interface{}{}
		if err != nil {
			payload["error"] = err.Error()
		}
		if stdout != "" {
			payload["output"] = stdout
		}
		if stderr != "" {
			payload["stderr"] = stderr
		}
		var data []byte
		if jsonPretty && !jsonCompact {
			if b, mErr := json.MarshalIndent(payload, "", "  "); mErr == nil {
				data = b
			} else {
				data, _ = json.Marshal(payload)
			}
		} else {
			data, _ = json.Marshal(payload)
		}
		return string(data), err
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

// SetupTestConfig sets up a minimal test configuration
func (h *TestHarness) SetupTestConfig() {
	// Create a temporary config for testing
	config.PrettyPrintConfig = false

	// Set up minimal test environment
	os.Setenv("WASP_CLI_TEST_MODE", "true")
}

// CleanupTestConfig cleans up test configuration
func (h *TestHarness) CleanupTestConfig() {
	os.Unsetenv("WASP_CLI_TEST_MODE")
}

// AssertOutputContains checks that the output contains the expected string
func (h *TestHarness) AssertOutputContains(output, expected string) {
	require.Contains(h.t, output, expected, "Output should contain: %s", expected)
}

// AssertOutputNotContains checks that the output does not contain the string
func (h *TestHarness) AssertOutputNotContains(output, unexpected string) {
	require.NotContains(h.t, output, unexpected, "Output should not contain: %s", unexpected)
}

// AssertOutputLines checks that the output has the expected number of lines
func (h *TestHarness) AssertOutputLines(output string, expectedLines int) {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) == 1 && lines[0] == "" {
		lines = []string{} // Empty output
	}
	require.Len(h.t, lines, expectedLines, "Expected %d lines, got %d", expectedLines, len(lines))
}

// GetOutputLines returns the output split into lines
func (h *TestHarness) GetOutputLines(output string) []string {
	if strings.TrimSpace(output) == "" {
		return []string{}
	}
	return strings.Split(strings.TrimSpace(output), "\n")
}

// WithContext runs a function with a test context
func (h *TestHarness) WithContext(fn func(ctx context.Context)) {
	ctx := context.Background()
	fn(ctx)
}

// JSON validation helpers

// IsValidJSON checks if the output is valid JSON
func (h *TestHarness) IsValidJSON(output string) bool {
	var js json.RawMessage
	return json.Unmarshal([]byte(output), &js) == nil
}

// AssertValidJSON checks that the output is valid JSON
func (h *TestHarness) AssertValidJSON(output string) {
	require.True(h.t, h.IsValidJSON(output), "Output should be valid JSON: %s", output)
}

// AssertJSONContains checks that the JSON output contains a specific key-value pair
func (h *TestHarness) AssertJSONContains(output, key, expectedValue string) {
	h.AssertValidJSON(output)

	var data map[string]interface{}
	err := json.Unmarshal([]byte(output), &data)
	require.NoError(h.t, err, "Failed to parse JSON")

	value, exists := data[key]
	require.True(h.t, exists, "JSON should contain key: %s", key)
	require.Equal(h.t, expectedValue, value, "JSON key %s should have value %s", key, expectedValue)
}

// AssertJSONHasKey checks that the JSON output contains a specific key
func (h *TestHarness) AssertJSONHasKey(output, key string) {
	h.AssertValidJSON(output)

	var data map[string]interface{}
	err := json.Unmarshal([]byte(output), &data)
	require.NoError(h.t, err, "Failed to parse JSON")

	_, exists := data[key]
	require.True(h.t, exists, "JSON should contain key: %s", key)
}

// AssertJSONNotHasKey checks that the JSON output does not contain a specific key
func (h *TestHarness) AssertJSONNotHasKey(output, key string) {
	h.AssertValidJSON(output)

	var data map[string]interface{}
	err := json.Unmarshal([]byte(output), &data)
	require.NoError(h.t, err, "Failed to parse JSON")

	_, exists := data[key]
	require.False(h.t, exists, "JSON should not contain key: %s", key)
}

// GetJSONValue extracts a value from JSON output by key
func (h *TestHarness) GetJSONValue(output, key string) interface{} {
	h.AssertValidJSON(output)

	var data map[string]interface{}
	err := json.Unmarshal([]byte(output), &data)
	require.NoError(h.t, err, "Failed to parse JSON")

	value, exists := data[key]
	require.True(h.t, exists, "JSON should contain key: %s", key)
	return value
}

// AssertJSONStructure checks that the JSON has the expected structure
func (h *TestHarness) AssertJSONStructure(output string, expectedKeys []string) {
	h.AssertValidJSON(output)

	var data map[string]interface{}
	err := json.Unmarshal([]byte(output), &data)
	require.NoError(h.t, err, "Failed to parse JSON")

	for _, key := range expectedKeys {
		_, exists := data[key]
		require.True(h.t, exists, "JSON should contain key: %s", key)
	}
}

// IsCompactJSON checks if JSON output is compact (no indentation)
func (h *TestHarness) IsCompactJSON(output string) bool {
	if !h.IsValidJSON(output) {
		return false
	}

	// Compact JSON should not contain indentation spaces or newlines within the JSON
	trimmed := strings.TrimSpace(output)
	return !strings.Contains(trimmed, "\n  ") && !strings.Contains(trimmed, "\t")
}

// IsPrettyJSON checks if JSON output is pretty-printed (with indentation)
func (h *TestHarness) IsPrettyJSON(output string) bool {
	if !h.IsValidJSON(output) {
		return false
	}

	// Pretty JSON should contain indentation
	return strings.Contains(output, "\n  ") || strings.Contains(output, "\n\t")
}

// AssertCompactJSON checks that the output is compact JSON
func (h *TestHarness) AssertCompactJSON(output string) {
	h.AssertValidJSON(output)
	require.True(h.t, h.IsCompactJSON(output), "Output should be compact JSON: %s", output)
}

// AssertPrettyJSON checks that the output is pretty-printed JSON
func (h *TestHarness) AssertPrettyJSON(output string) {
	h.AssertValidJSON(output)
	require.True(h.t, h.IsPrettyJSON(output), "Output should be pretty-printed JSON: %s", output)
}

// RunCommandWithJSON runs a command with the --json flag
func (h *TestHarness) RunCommandWithJSON(args ...string) (string, error) {
	jsonArgs := append(args, "--json")
	return h.RunCommand(jsonArgs...)
}

// MustRunCommandWithJSON runs a command with the --json flag and requires it to succeed
func (h *TestHarness) MustRunCommandWithJSON(args ...string) string {
	output, err := h.RunCommandWithJSON(args...)
	require.NoError(h.t, err, "Command with --json failed: %v", args)
	return output
}

// RunCommandWithJSONCompact runs a command with the --json-compact flag
func (h *TestHarness) RunCommandWithJSONCompact(args ...string) (string, error) {
	jsonArgs := append(args, "--json-compact")
	return h.RunCommand(jsonArgs...)
}

// MustRunCommandWithJSONCompact runs a command with the --json-compact flag and requires it to succeed
func (h *TestHarness) MustRunCommandWithJSONCompact(args ...string) string {
	output, err := h.RunCommandWithJSONCompact(args...)
	require.NoError(h.t, err, "Command with --json-compact failed: %v", args)
	return output
}
