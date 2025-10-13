package test

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestSimpleExample demonstrates the basic testing approach
func TestSimpleExample(t *testing.T) {
	h := NewTestHarness(t)
	h.SetupTestConfig()
	defer h.CleanupTestConfig()

	// Test that we can create a harness and run basic commands
	t.Run("basic help command", func(t *testing.T) {
		output := h.MustRunCommand("--help")
		require.NotEmpty(t, output)
		require.Contains(t, output, "wasp-cli")
	})
}

// TestMockBasics demonstrates mock functionality
func TestMockBasics(t *testing.T) {
	// Test L1Client mock
	l1Mock := NewMockL1Client()
	require.Equal(t, uint64(1000000), l1Mock.Balance)

	l1Mock.SetBalance(500000)
	require.Equal(t, uint64(500000), l1Mock.Balance)

	// Test Wallet mock
	walletMock := NewMockWallet()
	addr := walletMock.Address()
	require.NotNil(t, addr)
	require.True(t, walletMock.AddressCalled)

	// Test Chain Service mock
	chainMock := NewMockChainService()
	require.Empty(t, chainMock.ChainInfo)

	info := map[string]interface{}{"test": "value"}
	chainMock.SetChainInfo(info)
	require.Equal(t, info, chainMock.ChainInfo)
}

// TestFixturesBasics demonstrates test fixtures
func TestFixturesBasics(t *testing.T) {
	fixtures := NewTestFixtures()

	require.NotNil(t, fixtures.TestAddress1)
	require.NotNil(t, fixtures.TestAddress2)
	require.NotEqual(t, fixtures.TestAddress1, fixtures.TestAddress2)

	wallet := fixtures.GetTestWallet()
	require.NotNil(t, wallet)
}
