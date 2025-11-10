package test

import (
	"context"
	"errors"
	"math/big"
)

// MockL1Client provides a mock implementation of L1Client for testing
type MockL1Client struct {
	// Mock data
	Balance uint64
	Error   error

	// Call tracking
	GetBalanceCalled      bool
	GetCoinObjsCalled     bool
	DeployContractsCalled bool
}

// NewMockL1Client creates a new mock L1 client
func NewMockL1Client() *MockL1Client {
	return &MockL1Client{
		Balance: 1000000, // Default balance
	}
}

// GetBalance mocks getting the balance
func (m *MockL1Client) GetBalance(ctx context.Context, address string) (uint64, error) {
	m.GetBalanceCalled = true
	if m.Error != nil {
		return 0, m.Error
	}
	return m.Balance, nil
}

// GetCoinObjects mocks getting coin objects
func (m *MockL1Client) GetCoinObjects(ctx context.Context, address string, targetAmount, maxAmount uint64) ([]string, error) {
	m.GetCoinObjsCalled = true
	if m.Error != nil {
		return nil, m.Error
	}
	return []string{"coin1", "coin2"}, nil
}

// DeployContracts mocks deploying contracts
func (m *MockL1Client) DeployContracts(ctx context.Context, signer string) (string, error) {
	m.DeployContractsCalled = true
	if m.Error != nil {
		return "", m.Error
	}
	return "package-id-123", nil
}

// SetBalance sets the mock balance
func (m *MockL1Client) SetBalance(balance uint64) {
	m.Balance = balance
}

// SetError sets an error to be returned by mock methods
func (m *MockL1Client) SetError(err error) {
	m.Error = err
}

// Reset resets the mock state
func (m *MockL1Client) Reset() {
	m.GetBalanceCalled = false
	m.GetCoinObjsCalled = false
	m.DeployContractsCalled = false
	m.Error = nil
}

// MockWallet provides a mock implementation of Wallet for testing
type MockWallet struct {
	// Mock data
	WalletAddress string
	Error         error

	// Call tracking
	AddressCalled bool
	SignCalled    bool
}

// NewMockWallet creates a new mock wallet
func NewMockWallet() *MockWallet {
	return &MockWallet{
		WalletAddress: "test-wallet-address-123",
	}
}

// Address returns the mock wallet address
func (m *MockWallet) Address() string {
	m.AddressCalled = true
	return m.WalletAddress
}

// Sign mocks signing data
func (m *MockWallet) Sign(data []byte) ([]byte, error) {
	m.SignCalled = true
	if m.Error != nil {
		return nil, m.Error
	}
	return []byte("mock-signature"), nil
}

// SetAddress sets the mock address
func (m *MockWallet) SetAddress(address string) {
	m.WalletAddress = address
}

// SetError sets an error to be returned by mock methods
func (m *MockWallet) SetError(err error) {
	m.Error = err
}

// Reset resets the mock state
func (m *MockWallet) Reset() {
	m.AddressCalled = false
	m.SignCalled = false
	m.Error = nil
}

// MockChainService provides a mock implementation of chain service for testing
type MockChainService struct {
	// Mock data
	ChainID   string
	ChainInfo map[string]interface{}
	Balance   map[string]*big.Int
	Error     error

	// Call tracking
	GetChainInfoCalled bool
	GetBalanceCalled   bool
	DeployChainCalled  bool
}

// NewMockChainService creates a new mock chain service
func NewMockChainService() *MockChainService {
	return &MockChainService{
		ChainInfo: make(map[string]interface{}),
		Balance:   make(map[string]*big.Int),
		ChainID:   "test-chain-id-123",
	}
}

// GetChainInfo mocks getting chain information
func (m *MockChainService) GetChainInfo(ctx context.Context) (map[string]interface{}, error) {
	m.GetChainInfoCalled = true
	if m.Error != nil {
		return nil, m.Error
	}
	return m.ChainInfo, nil
}

// GetBalance mocks getting account balance
func (m *MockChainService) GetBalance(ctx context.Context, address string) (map[string]*big.Int, error) {
	m.GetBalanceCalled = true
	if m.Error != nil {
		return nil, m.Error
	}
	return m.Balance, nil
}

// DeployChain mocks deploying a chain
func (m *MockChainService) DeployChain(ctx context.Context, params interface{}) (string, error) {
	m.DeployChainCalled = true
	if m.Error != nil {
		return "", m.Error
	}
	return m.ChainID, nil
}

// SetChainInfo sets mock chain info
func (m *MockChainService) SetChainInfo(info map[string]interface{}) {
	m.ChainInfo = info
}

// SetBalance sets mock balance
func (m *MockChainService) SetBalance(tokenType string, amount *big.Int) {
	m.Balance[tokenType] = amount
}

// SetChainID sets mock chain ID
func (m *MockChainService) SetChainID(chainID string) {
	m.ChainID = chainID
}

// SetError sets an error to be returned by mock methods
func (m *MockChainService) SetError(err error) {
	m.Error = err
}

// Reset resets the mock state
func (m *MockChainService) Reset() {
	m.GetChainInfoCalled = false
	m.GetBalanceCalled = false
	m.DeployChainCalled = false
	m.Error = nil
	m.ChainInfo = make(map[string]interface{})
	m.Balance = make(map[string]*big.Int)
}

// TestFixtures provides common test data and utilities
type TestFixtures struct {
	// Common test addresses
	TestAddress1 string
	TestAddress2 string

	// Common test chain IDs
	TestChainID1 string
	TestChainID2 string
}

// NewTestFixtures creates common test fixtures
func NewTestFixtures() *TestFixtures {
	return &TestFixtures{
		TestAddress1: "test-address-1",
		TestAddress2: "test-address-2",
		TestChainID1: "test-chain-1",
		TestChainID2: "test-chain-2",
	}
}

// GetTestWallet returns a test wallet implementation
func (f *TestFixtures) GetTestWallet() *MockWallet {
	wallet := NewMockWallet()
	wallet.SetAddress(f.TestAddress1)
	return wallet
}

// MockConfig provides a mock configuration for testing
type MockConfig struct {
	Values map[string]interface{}
	Error  error
}

// NewMockConfig creates a new mock config
func NewMockConfig() *MockConfig {
	return &MockConfig{
		Values: make(map[string]interface{}),
	}
}

// Get mocks getting a config value
func (m *MockConfig) Get(key string) (interface{}, error) {
	if m.Error != nil {
		return nil, m.Error
	}
	return m.Values[key], nil
}

// Set mocks setting a config value
func (m *MockConfig) Set(key string, value interface{}) error {
	if m.Error != nil {
		return m.Error
	}
	m.Values[key] = value
	return nil
}

// SetValue sets a mock config value
func (m *MockConfig) SetValue(key string, value interface{}) {
	m.Values[key] = value
}

// SetError sets an error to be returned
func (m *MockConfig) SetError(err error) {
	m.Error = err
}

// Reset resets the mock state
func (m *MockConfig) Reset() {
	m.Values = make(map[string]interface{})
	m.Error = nil
}

// Example usage in tests:
//
// func TestWithMocks(t *testing.T) {
//     l1Mock := NewMockL1Client()
//     l1Mock.SetBalance(500000)
//
//     balance, err := l1Mock.GetBalance(context.Background(), "test-address")
//     require.NoError(t, err)
//     require.Equal(t, uint64(500000), balance)
//     require.True(t, l1Mock.GetBalanceCalled)
// }

// TestErrorScenario demonstrates error testing
func TestErrorScenario() error {
	mock := NewMockL1Client()
	mock.SetError(errors.New("test error"))

	_, err := mock.GetBalance(context.Background(), "test-address")
	return err // Should return the test error
}
