package test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestSimpleMockL1Client tests the L1 client mock
func TestSimpleMockL1Client(t *testing.T) {
	mock := NewSimpleMockL1Client()

	t.Run("default values", func(t *testing.T) {
		require.Equal(t, uint64(1000000), mock.Balance)
		require.False(t, mock.GetBalanceCalled)
	})

	t.Run("get balance success", func(t *testing.T) {
		balance, err := mock.GetBalance(context.Background(), "test-address")
		require.NoError(t, err)
		require.Equal(t, uint64(1000000), balance)
		require.True(t, mock.GetBalanceCalled)
	})

	t.Run("set balance", func(t *testing.T) {
		mock.SetBalance(500000)
		require.Equal(t, uint64(500000), mock.Balance)

		balance, err := mock.GetBalance(context.Background(), "test-address")
		require.NoError(t, err)
		require.Equal(t, uint64(500000), balance)
	})

	t.Run("error handling", func(t *testing.T) {
		mock.Reset()
		testError := errors.New("test error")
		mock.SetError(testError)

		_, err := mock.GetBalance(context.Background(), "test-address")
		require.Error(t, err)
		require.Equal(t, testError, err)
		require.True(t, mock.GetBalanceCalled)
	})

	t.Run("reset", func(t *testing.T) {
		mock.GetBalanceCalled = true
		mock.SetError(errors.New("test"))

		mock.Reset()
		require.False(t, mock.GetBalanceCalled)
		require.Nil(t, mock.Error)
	})
}

// TestSimpleMockWallet tests the wallet mock
func TestSimpleMockWallet(t *testing.T) {
	mock := NewSimpleMockWallet()

	t.Run("default values", func(t *testing.T) {
		require.Equal(t, "test-address-123", mock.Address)
		require.False(t, mock.AddressCalled)
	})

	t.Run("get address success", func(t *testing.T) {
		address, err := mock.GetAddress()
		require.NoError(t, err)
		require.Equal(t, "test-address-123", address)
		require.True(t, mock.AddressCalled)
	})

	t.Run("set address", func(t *testing.T) {
		mock.SetAddress("new-address")
		require.Equal(t, "new-address", mock.Address)

		address, err := mock.GetAddress()
		require.NoError(t, err)
		require.Equal(t, "new-address", address)
	})

	t.Run("error handling", func(t *testing.T) {
		mock.Reset()
		testError := errors.New("wallet error")
		mock.SetError(testError)

		_, err := mock.GetAddress()
		require.Error(t, err)
		require.Equal(t, testError, err)
		require.True(t, mock.AddressCalled)
	})
}

// TestSimpleMockChainService tests the chain service mock
func TestSimpleMockChainService(t *testing.T) {
	mock := NewSimpleMockChainService()

	t.Run("default values", func(t *testing.T) {
		require.Empty(t, mock.ChainInfo)
		require.False(t, mock.GetChainInfoCalled)
	})

	t.Run("get chain info success", func(t *testing.T) {
		info, err := mock.GetChainInfo(context.Background())
		require.NoError(t, err)
		require.Empty(t, info)
		require.True(t, mock.GetChainInfoCalled)
	})

	t.Run("set chain info", func(t *testing.T) {
		testInfo := map[string]interface{}{
			"chainId": "test-chain",
			"status":  "active",
		}
		mock.SetChainInfo(testInfo)
		require.Equal(t, testInfo, mock.ChainInfo)

		info, err := mock.GetChainInfo(context.Background())
		require.NoError(t, err)
		require.Equal(t, testInfo, info)
	})

	t.Run("error handling", func(t *testing.T) {
		mock.Reset()
		testError := errors.New("chain error")
		mock.SetError(testError)

		_, err := mock.GetChainInfo(context.Background())
		require.Error(t, err)
		require.Equal(t, testError, err)
		require.True(t, mock.GetChainInfoCalled)
	})
}

// TestSimpleTestFixtures tests the test fixtures
func TestSimpleTestFixtures(t *testing.T) {
	fixtures := NewSimpleTestFixtures()

	t.Run("fixture values", func(t *testing.T) {
		require.Equal(t, "test-address-1", fixtures.TestAddress1)
		require.Equal(t, "test-address-2", fixtures.TestAddress2)
		require.Equal(t, "test-chain-1", fixtures.TestChainID1)
		require.Equal(t, "test-chain-2", fixtures.TestChainID2)

		require.NotEqual(t, fixtures.TestAddress1, fixtures.TestAddress2)
		require.NotEqual(t, fixtures.TestChainID1, fixtures.TestChainID2)
	})

	t.Run("get test wallet", func(t *testing.T) {
		wallet := fixtures.GetTestWallet()
		require.NotNil(t, wallet)
		require.Equal(t, fixtures.TestAddress1, wallet.Address)
	})
}

// TestMockService tests the generic mock service
func TestMockService(t *testing.T) {
	mock := NewMockService()

	t.Run("default values", func(t *testing.T) {
		require.Equal(t, "default-value", mock.Value)
		require.False(t, mock.Called)
	})

	t.Run("do something success", func(t *testing.T) {
		value, err := mock.DoSomething()
		require.NoError(t, err)
		require.Equal(t, "default-value", value)
		require.True(t, mock.Called)
	})

	t.Run("set value", func(t *testing.T) {
		mock.SetValue("custom-value")
		require.Equal(t, "custom-value", mock.Value)

		value, err := mock.DoSomething()
		require.NoError(t, err)
		require.Equal(t, "custom-value", value)
	})

	t.Run("error handling", func(t *testing.T) {
		mock.Reset()
		testError := errors.New("service error")
		mock.SetError(testError)

		_, err := mock.DoSomething()
		require.Error(t, err)
		require.Equal(t, testError, err)
		require.True(t, mock.Called)
	})

	t.Run("reset", func(t *testing.T) {
		mock.Called = true
		mock.SetError(errors.New("test"))
		mock.SetValue("test")

		mock.Reset()
		require.False(t, mock.Called)
		require.Nil(t, mock.Error)
		require.Equal(t, "default-value", mock.Value)
	})
}

// TestMockIntegration demonstrates using multiple mocks together
func TestMockIntegration(t *testing.T) {
	l1Mock := NewSimpleMockL1Client()
	walletMock := NewSimpleMockWallet()
	chainMock := NewSimpleMockChainService()
	fixtures := NewSimpleTestFixtures()

	// Configure mocks
	l1Mock.SetBalance(2000000)
	walletMock.SetAddress(fixtures.TestAddress1)
	chainMock.SetChainInfo(map[string]interface{}{
		"chainId": fixtures.TestChainID1,
		"status":  "active",
	})

	// Test integration
	balance, err := l1Mock.GetBalance(context.Background(), fixtures.TestAddress1)
	require.NoError(t, err)
	require.Equal(t, uint64(2000000), balance)

	address, err := walletMock.GetAddress()
	require.NoError(t, err)
	require.Equal(t, fixtures.TestAddress1, address)

	chainInfo, err := chainMock.GetChainInfo(context.Background())
	require.NoError(t, err)
	require.Equal(t, fixtures.TestChainID1, chainInfo["chainId"])
	require.Equal(t, "active", chainInfo["status"])

	// Verify all mocks were called
	require.True(t, l1Mock.GetBalanceCalled)
	require.True(t, walletMock.AddressCalled)
	require.True(t, chainMock.GetChainInfoCalled)
}
