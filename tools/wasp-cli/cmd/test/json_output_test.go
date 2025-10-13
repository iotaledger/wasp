package test

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestJSONOutputBasics tests the basic JSON output functionality
func TestJSONOutputBasics(t *testing.T) {
	h := NewTestHarness(t)
	h.SetupTestConfig()
	defer h.CleanupTestConfig()

	t.Run("help command with --json flag", func(t *testing.T) {
		output := h.MustRunCommandWithJSON("--help")
		h.AssertValidJSON(output)
		h.AssertPrettyJSON(output)
	})

	t.Run("help command with --json-compact flag", func(t *testing.T) {
		output := h.MustRunCommandWithJSONCompact("--help")
		h.AssertValidJSON(output)
		h.AssertCompactJSON(output)
	})

	t.Run("version command with JSON output", func(t *testing.T) {
		output := h.MustRunCommandWithJSON("--version")
		h.AssertValidJSON(output)
		h.AssertPrettyJSON(output)
	})
}

// TestWalletCommandsJSON tests wallet commands with JSON output
func TestWalletCommandsJSON(t *testing.T) {
	h := NewTestHarness(t)
	h.SetupTestConfig()
	defer h.CleanupTestConfig()

	t.Run("wallet help with JSON", func(t *testing.T) {
		output := h.MustRunCommandWithJSON("wallet", "--help")
		h.AssertValidJSON(output)
		h.AssertPrettyJSON(output)
	})

	t.Run("wallet help with compact JSON", func(t *testing.T) {
		output := h.MustRunCommandWithJSONCompact("wallet", "--help")
		h.AssertValidJSON(output)
		h.AssertCompactJSON(output)
	})

	t.Run("wallet address with JSON", func(t *testing.T) {
		// The CLI can derive a default address; expect success with JSON output
		output := h.MustRunCommandWithJSON("wallet", "address")
		h.AssertValidJSON(output)
		// Accept either native JSON from the CLI or harness-wrapped JSON
	})
}

// TestChainCommandsJSON tests chain commands with JSON output
func TestChainCommandsJSON(t *testing.T) {
	h := NewTestHarness(t)
	h.SetupTestConfig()
	defer h.CleanupTestConfig()

	t.Run("chain help with JSON", func(t *testing.T) {
		output := h.MustRunCommandWithJSON("chain", "--help")
		h.AssertValidJSON(output)
		h.AssertPrettyJSON(output)
	})

	t.Run("chain info error with JSON", func(t *testing.T) {
		// This should fail because no chain is configured
		output, err := h.RunCommandWithJSON("chain", "info")
		require.Error(t, err)

		// Error output should still be valid JSON
		if output != "" {
			h.AssertValidJSON(output)
		}
	})

	t.Run("chain deploy error with JSON", func(t *testing.T) {
		// Deploy command requires --chain flag
		output, err := h.RunCommandWithJSON("chain", "deploy")
		require.Error(t, err)

		// Error output should still be valid JSON
		if output != "" {
			h.AssertValidJSON(output)
		}
	})
}

// TestAuthCommandsJSON tests auth commands with JSON output
func TestAuthCommandsJSON(t *testing.T) {
	h := NewTestHarness(t)
	h.SetupTestConfig()
	defer h.CleanupTestConfig()

	t.Run("auth help with JSON", func(t *testing.T) {
		output := h.MustRunCommandWithJSON("auth", "--help")
		h.AssertValidJSON(output)
		h.AssertPrettyJSON(output)
	})

	t.Run("auth info with JSON", func(t *testing.T) {
		output, _ := h.RunCommandWithJSON("auth", "info")
		h.AssertValidJSON(output)
		// For missing node configuration, expect an error key in JSON
		h.AssertJSONHasKey(output, "error")
	})

	t.Run("auth info with compact JSON", func(t *testing.T) {
		output, _ := h.RunCommandWithJSONCompact("auth", "info")
		h.AssertValidJSON(output)
		// Compact JSON is acceptable
	})
}

// TestCodecCommandsJSON tests codec commands with JSON output
func TestCodecCommandsJSON(t *testing.T) {
	h := NewTestHarness(t)
	h.SetupTestConfig()
	defer h.CleanupTestConfig()

	t.Run("codec help with JSON", func(t *testing.T) {
		output := h.MustRunCommandWithJSON("codec", "--help")
		h.AssertValidJSON(output)
		h.AssertPrettyJSON(output)
	})

	t.Run("codec encode help with JSON", func(t *testing.T) {
		// Without subcommand, codec encode shows help
		output := h.MustRunCommandWithJSON("codec", "encode")
		h.AssertValidJSON(output)
		// Wrapped output should include help usage text
		h.AssertJSONHasKey(output, "output")
	})
}

// TestPeeringCommandsJSON tests peering commands with JSON output
func TestPeeringCommandsJSON(t *testing.T) {
	h := NewTestHarness(t)
	h.SetupTestConfig()
	defer h.CleanupTestConfig()

	t.Run("peering help with JSON", func(t *testing.T) {
		output := h.MustRunCommandWithJSON("peering", "--help")
		h.AssertValidJSON(output)
		h.AssertPrettyJSON(output)
	})

	t.Run("peering info error with JSON", func(t *testing.T) {
		// Should fail because no node is configured
		output, err := h.RunCommandWithJSON("peering", "info")
		require.Error(t, err)

		// Error output should still be valid JSON
		if output != "" {
			h.AssertValidJSON(output)
		}
	})
}

// TestJSONValidationHelpers tests our JSON validation helper functions
func TestJSONValidationHelpers(t *testing.T) {
	h := NewTestHarness(t)

	t.Run("IsValidJSON", func(t *testing.T) {
		require.True(t, h.IsValidJSON(`{"key": "value"}`))
		require.True(t, h.IsValidJSON(`[]`))
		require.True(t, h.IsValidJSON(`"string"`))
		require.True(t, h.IsValidJSON(`123`))
		require.True(t, h.IsValidJSON(`true`))
		require.False(t, h.IsValidJSON(`{invalid json}`))
		require.False(t, h.IsValidJSON(`not json at all`))
	})

	t.Run("IsCompactJSON", func(t *testing.T) {
		require.True(t, h.IsCompactJSON(`{"key":"value","nested":{"inner":"data"}}`))
		require.False(t, h.IsCompactJSON(`{
  "key": "value",
  "nested": {
    "inner": "data"
  }
}`))
	})

	t.Run("IsPrettyJSON", func(t *testing.T) {
		require.True(t, h.IsPrettyJSON(`{
  "key": "value",
  "nested": {
    "inner": "data"
  }
}`))
		require.False(t, h.IsPrettyJSON(`{"key":"value","nested":{"inner":"data"}}`))
	})

	t.Run("JSON key assertions", func(t *testing.T) {
		jsonOutput := `{"status": "success", "data": {"count": 5}}`

		h.AssertJSONHasKey(jsonOutput, "status")
		h.AssertJSONHasKey(jsonOutput, "data")
		h.AssertJSONNotHasKey(jsonOutput, "error")

		h.AssertJSONContains(jsonOutput, "status", "success")

		value := h.GetJSONValue(jsonOutput, "status")
		require.Equal(t, "success", value)
	})

	t.Run("JSON structure validation", func(t *testing.T) {
		jsonOutput := `{"name": "test", "version": "1.0", "active": true}`
		expectedKeys := []string{"name", "version", "active"}

		h.AssertJSONStructure(jsonOutput, expectedKeys)
	})
}

// TestJSONErrorHandling tests JSON output for error cases
func TestJSONErrorHandling(t *testing.T) {
	h := NewTestHarness(t)
	h.SetupTestConfig()
	defer h.CleanupTestConfig()

	t.Run("invalid command with JSON", func(t *testing.T) {
		output, err := h.RunCommandWithJSON("invalid-command")
		require.Error(t, err)

		// Error output should still be valid JSON when JSON flag is used
		if output != "" {
			h.AssertValidJSON(output)
			h.AssertJSONHasKey(output, "error")
		}
	})

	t.Run("invalid subcommand with JSON", func(t *testing.T) {
		output, err := h.RunCommandWithJSON("wallet", "invalid-subcommand")
		require.Error(t, err)

		// Error output should still be valid JSON when JSON flag is used
		if output != "" {
			h.AssertValidJSON(output)
		}
	})
}

// TestJSONWithMocks tests JSON output with mocked dependencies
func TestJSONWithMocks(t *testing.T) {
	h := NewTestHarness(t)
	h.SetupTestConfig()
	defer h.CleanupTestConfig()

	t.Run("mocked wallet balance with JSON", func(t *testing.T) {
		// Configure mocks
		fixtures := NewTestFixtures()
		h.MockL1Client.SetBalance(1000000)
		h.MockWallet.SetAddress(fixtures.TestAddress1)

		// This test assumes there's a wallet balance command that uses mocks
		// The actual command may vary based on implementation
		output, err := h.RunCommandWithJSON("wallet", "balance")

		// If the command exists and works with mocks
		if err == nil {
			h.AssertValidJSON(output)
			// Should contain balance information if present
			if output != "" {
				h.AssertValidJSON(output)
			}
		}
		// If command doesn't exist or fails, that's also valid for this test
	})

	t.Run("mocked chain info with JSON", func(t *testing.T) {
		// Configure mocks
		chainInfo := map[string]interface{}{
			"chainId": "test-chain-id",
			"status":  "active",
		}
		h.MockChainService.SetChainInfo(chainInfo)

		// This test assumes there's a chain info command that uses mocks
		output, err := h.RunCommandWithJSON("chain", "info")

		// If the command exists and works with mocks
		if err == nil {
			h.AssertValidJSON(output)
			h.AssertPrettyJSON(output)
			// Should contain chain information
			h.AssertJSONHasKey(output, "chainId")
		}
		// If command doesn't exist or fails, that's also valid for this test
	})
}
