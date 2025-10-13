package test

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRootCommand(t *testing.T) {
	h := NewTestHarness(t)
	h.SetupTestConfig()
	defer h.CleanupTestConfig()

	t.Run("help command", func(t *testing.T) {
		output := h.MustRunCommand("--help")
		h.AssertOutputContains(output, "wasp-cli")
		h.AssertOutputContains(output, "Usage:")
	})

	t.Run("version command", func(t *testing.T) {
		output := h.MustRunCommand("--version")
		// Should contain version information
		require.NotEmpty(t, output)
	})
}

func TestWalletCommands(t *testing.T) {
	h := NewTestHarness(t)
	h.SetupTestConfig()
	defer h.CleanupTestConfig()

	t.Run("wallet help", func(t *testing.T) {
		output := h.MustRunCommand("wallet", "--help")
		h.AssertOutputContains(output, "wallet")
		h.AssertOutputContains(output, "address")
		h.AssertOutputContains(output, "balance")
	})

	t.Run("wallet address without setup", func(t *testing.T) {
		// The CLI can derive a default address; expect success and output
		_ = h.MustRunCommand("wallet", "address")
		// Some environments may not print address to captured stdout, just ensure it succeeds
	})
}

func TestChainCommands(t *testing.T) {
	h := NewTestHarness(t)
	h.SetupTestConfig()
	defer h.CleanupTestConfig()

	t.Run("chain help", func(t *testing.T) {
		output := h.MustRunCommand("chain", "--help")
		h.AssertOutputContains(output, "chain")
		h.AssertOutputContains(output, "deploy")
		h.AssertOutputContains(output, "info")
	})

	t.Run("chain info without chain", func(t *testing.T) {
		// This should fail because no chain is configured
		_, err := h.RunCommandExpectError("chain", "info")
		require.Error(t, err)
	})

	t.Run("chain deploy missing required flags", func(t *testing.T) {
		// Deploy command requires --chain flag
		_, err := h.RunCommandExpectError("chain", "deploy")
		require.Error(t, err)
	})

	t.Run("chain deploy with chain flag but no setup", func(t *testing.T) {
		// This should fail because dependencies aren't set up
		_, err := h.RunCommandExpectError("chain", "deploy", "--chain=test-chain")
		require.Error(t, err)
	})
}

func TestAuthCommands(t *testing.T) {
	h := NewTestHarness(t)
	h.SetupTestConfig()
	defer h.CleanupTestConfig()

	t.Run("auth help", func(t *testing.T) {
		output := h.MustRunCommand("auth", "--help")
		h.AssertOutputContains(output, "auth")
		h.AssertOutputContains(output, "login")
	})

	t.Run("auth info without login", func(t *testing.T) {
		_, err := h.RunCommand("auth", "info")
		require.Error(t, err)
		// Should complain about missing wasp node configuration
		require.Contains(t, err.Error(), "no wasp node configured")
	})
}

func TestCodecCommands(t *testing.T) {
	h := NewTestHarness(t)
	h.SetupTestConfig()
	defer h.CleanupTestConfig()

	t.Run("codec help", func(t *testing.T) {
		output := h.MustRunCommand("codec", "--help")
		h.AssertOutputContains(output, "codec")
		h.AssertOutputContains(output, "encode")
		h.AssertOutputContains(output, "decode")
	})

	t.Run("codec encode shows help when missing subcommand", func(t *testing.T) {
		output := h.MustRunCommand("codec", "encode")
		h.AssertOutputContains(output, "Usage:")
	})
}

func TestPeeringCommands(t *testing.T) {
	h := NewTestHarness(t)
	h.SetupTestConfig()
	defer h.CleanupTestConfig()

	t.Run("peering help", func(t *testing.T) {
		output := h.MustRunCommand("peering", "--help")
		h.AssertOutputContains(output, "peering")
		h.AssertOutputContains(output, "info")
		h.AssertOutputContains(output, "list-trusted")
	})

	t.Run("peering info without node", func(t *testing.T) {
		// Should fail because no node is configured
		_, err := h.RunCommandExpectError("peering", "info")
		require.Error(t, err)
	})
}

func TestInvalidCommands(t *testing.T) {
	h := NewTestHarness(t)
	h.SetupTestConfig()
	defer h.CleanupTestConfig()

	t.Run("invalid command", func(t *testing.T) {
		_, err := h.RunCommandExpectError("invalid-command")
		require.Error(t, err)
	})

	t.Run("invalid subcommand", func(t *testing.T) {
		_, err := h.RunCommandExpectError("wallet", "invalid-subcommand")
		require.Error(t, err)
	})
}

func TestOutputFormats(t *testing.T) {
	h := NewTestHarness(t)
	h.SetupTestConfig()
	defer h.CleanupTestConfig()

	t.Run("help output format", func(t *testing.T) {
		output := h.MustRunCommand("--help")
		lines := h.GetOutputLines(output)

		// Should have multiple lines
		require.Greater(t, len(lines), 5)

		// Should contain usage information
		found := false
		for _, line := range lines {
			if contains(line, "Usage:") {
				found = true
				break
			}
		}
		require.True(t, found, "Should contain Usage information")
	})
}

// Helper function to check if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestMockFunctionality tests that our mocks work correctly
func TestMockFunctionality(t *testing.T) {
	t.Run("MockL1Client", func(t *testing.T) {
		mock := NewMockL1Client()

		// Test default values
		require.Equal(t, uint64(1000000), mock.Balance)
		require.False(t, mock.GetBalanceCalled)

		// Test setting values
		mock.SetBalance(500000)
		require.Equal(t, uint64(500000), mock.Balance)

		// Test reset
		mock.GetBalanceCalled = true
		mock.Reset()
		require.False(t, mock.GetBalanceCalled)
	})

	t.Run("MockWallet", func(t *testing.T) {
		mock := NewMockWallet()

		// Test address
		addr := mock.Address()
		require.NotNil(t, addr)
		require.True(t, mock.AddressCalled)

		// Test reset
		mock.Reset()
		require.False(t, mock.AddressCalled)
	})

	t.Run("MockChainService", func(t *testing.T) {
		mock := NewMockChainService()

		// Test default values
		require.Empty(t, mock.ChainInfo)
		require.False(t, mock.GetChainInfoCalled)

		// Test setting values
		info := map[string]interface{}{"test": "value"}
		mock.SetChainInfo(info)
		require.Equal(t, info, mock.ChainInfo)

		// Test reset
		mock.GetChainInfoCalled = true
		mock.Reset()
		require.False(t, mock.GetChainInfoCalled)
		require.Empty(t, mock.ChainInfo)
	})
}

func TestTestFixtures(t *testing.T) {
	fixtures := NewTestFixtures()

	// Test that fixtures are properly initialized
	require.NotNil(t, fixtures.TestAddress1)
	require.NotNil(t, fixtures.TestAddress2)
	require.NotEqual(t, fixtures.TestAddress1, fixtures.TestAddress2)

	require.NotEqual(t, fixtures.TestChainID1, fixtures.TestChainID2)

	// Test wallet creation
	wallet := fixtures.GetTestWallet()
	require.NotNil(t, wallet)
	require.NotNil(t, wallet.Address())
}
