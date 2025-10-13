package test

import (
	"context"
	"errors"
)

// SimpleMockL1Client provides a basic mock for L1 operations
type SimpleMockL1Client struct {
	Balance          uint64
	Error            error
	GetBalanceCalled bool
}

// NewSimpleMockL1Client creates a new simple L1 client mock
func NewSimpleMockL1Client() *SimpleMockL1Client {
	return &SimpleMockL1Client{
		Balance: 1000000, // Default balance
	}
}

// GetBalance mocks getting the balance
func (m *SimpleMockL1Client) GetBalance(ctx context.Context, address string) (uint64, error) {
	m.GetBalanceCalled = true
	if m.Error != nil {
		return 0, m.Error
	}
	return m.Balance, nil
}

// SetBalance sets the mock balance
func (m *SimpleMockL1Client) SetBalance(balance uint64) {
	m.Balance = balance
}

// SetError sets an error to be returned by mock methods
func (m *SimpleMockL1Client) SetError(err error) {
	m.Error = err
}

// Reset resets the mock state
func (m *SimpleMockL1Client) Reset() {
	m.GetBalanceCalled = false
	m.Error = nil
}

// SimpleMockWallet provides a basic mock for wallet operations
type SimpleMockWallet struct {
	Address       string
	Error         error
	AddressCalled bool
}

// NewSimpleMockWallet creates a new simple wallet mock
func NewSimpleMockWallet() *SimpleMockWallet {
	return &SimpleMockWallet{
		Address: "test-address-123",
	}
}

// GetAddress returns the mock wallet address
func (m *SimpleMockWallet) GetAddress() (string, error) {
	m.AddressCalled = true
	if m.Error != nil {
		return "", m.Error
	}
	return m.Address, nil
}

// SetAddress sets the mock address
func (m *SimpleMockWallet) SetAddress(address string) {
	m.Address = address
}

// SetError sets an error to be returned by mock methods
func (m *SimpleMockWallet) SetError(err error) {
	m.Error = err
}

// Reset resets the mock state
func (m *SimpleMockWallet) Reset() {
	m.AddressCalled = false
	m.Error = nil
}

// SimpleMockChainService provides a basic mock for chain operations
type SimpleMockChainService struct {
	ChainInfo          map[string]interface{}
	Error              error
	GetChainInfoCalled bool
}

// NewSimpleMockChainService creates a new simple chain service mock
func NewSimpleMockChainService() *SimpleMockChainService {
	return &SimpleMockChainService{
		ChainInfo: make(map[string]interface{}),
	}
}

// GetChainInfo mocks getting chain information
func (m *SimpleMockChainService) GetChainInfo(ctx context.Context) (map[string]interface{}, error) {
	m.GetChainInfoCalled = true
	if m.Error != nil {
		return nil, m.Error
	}
	return m.ChainInfo, nil
}

// SetChainInfo sets mock chain info
func (m *SimpleMockChainService) SetChainInfo(info map[string]interface{}) {
	m.ChainInfo = info
}

// SetError sets an error to be returned by mock methods
func (m *SimpleMockChainService) SetError(err error) {
	m.Error = err
}

// Reset resets the mock state
func (m *SimpleMockChainService) Reset() {
	m.GetChainInfoCalled = false
	m.Error = nil
	m.ChainInfo = make(map[string]interface{})
}

// SimpleTestFixtures provides basic test data
type SimpleTestFixtures struct {
	TestAddress1 string
	TestAddress2 string
	TestChainID1 string
	TestChainID2 string
}

// NewSimpleTestFixtures creates simple test fixtures
func NewSimpleTestFixtures() *SimpleTestFixtures {
	return &SimpleTestFixtures{
		TestAddress1: "test-address-1",
		TestAddress2: "test-address-2",
		TestChainID1: "test-chain-1",
		TestChainID2: "test-chain-2",
	}
}

// GetTestWallet returns a simple test wallet
func (f *SimpleTestFixtures) GetTestWallet() *SimpleMockWallet {
	wallet := NewSimpleMockWallet()
	wallet.SetAddress(f.TestAddress1)
	return wallet
}

// MockService demonstrates a generic mock pattern
type MockService struct {
	Value  string
	Called bool
	Error  error
}

// NewMockService creates a new mock service
func NewMockService() *MockService {
	return &MockService{
		Value: "default-value",
	}
}

// DoSomething mocks a service operation
func (m *MockService) DoSomething() (string, error) {
	m.Called = true
	if m.Error != nil {
		return "", m.Error
	}
	return m.Value, nil
}

// SetValue sets the mock return value
func (m *MockService) SetValue(value string) {
	m.Value = value
}

// SetError sets an error to be returned
func (m *MockService) SetError(err error) {
	m.Error = err
}

// Reset resets the mock state
func (m *MockService) Reset() {
	m.Called = false
	m.Error = nil
	m.Value = "default-value"
}

// Example of how to use these mocks in tests:
//
// func TestWithMocks(t *testing.T) {
//     l1Mock := NewSimpleMockL1Client()
//     l1Mock.SetBalance(500000)
//
//     balance, err := l1Mock.GetBalance(context.Background(), "test-address")
//     require.NoError(t, err)
//     require.Equal(t, uint64(500000), balance)
//     require.True(t, l1Mock.GetBalanceCalled)
// }

// TestMockError demonstrates error testing
func TestMockError() error {
	mock := NewMockService()
	mock.SetError(errors.New("test error"))

	_, err := mock.DoSomething()
	return err // Should return the test error
}
