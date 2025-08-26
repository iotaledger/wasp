package format

import (
	"fmt"
	"strconv"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotajsonrpc"
)

// WalletBalanceData represents the data structure for wallet balance commands
type WalletBalanceData struct {
	AddressIndex uint32                 `json:"address_index"`
	Address      string                 `json:"address"`
	Balances     []*iotajsonrpc.Balance `json:"balances"`
}

// WalletBalanceOutput represents the output of wallet balance commands
type WalletBalanceOutput struct {
	BaseOutput
	WalletBalanceData WalletBalanceData `json:"-"` // Embedded for easy access, but not serialized directly
}

// NewWalletBalanceOutput creates a new wallet balance output
func NewWalletBalanceOutput(addressIndex uint32, address string, balances []*iotajsonrpc.Balance, success bool) *WalletBalanceOutput {
	status := "success"
	if !success {
		status = "error"
	}

	walletBalanceData := WalletBalanceData{
		AddressIndex: addressIndex,
		Address:      address,
		Balances:     balances,
	}

	return &WalletBalanceOutput{
		BaseOutput:        NewBaseOutput("wallet_balance", status, walletBalanceData),
		WalletBalanceData: walletBalanceData,
	}
}

// NewWalletBalanceSuccess creates a successful wallet balance output
func NewWalletBalanceSuccess(addressIndex uint32, address string, balances []*iotajsonrpc.Balance) *WalletBalanceOutput {
	return NewWalletBalanceOutput(addressIndex, address, balances, true)
}

// NewWalletBalanceError creates an error wallet balance output
func NewWalletBalanceError(addressIndex uint32, address string, errorMsg string) *WalletBalanceOutput {
	// Create empty balances for error case
	return &WalletBalanceOutput{
		BaseOutput: NewBaseOutput("wallet_balance", "error", map[string]interface{}{
			"address_index": addressIndex,
			"address":       address,
			"error":         errorMsg,
		}),
		WalletBalanceData: WalletBalanceData{
			AddressIndex: addressIndex,
			Address:      address,
			Balances:     []*iotajsonrpc.Balance{},
		},
	}
}

// ToJSON returns the wallet balance output as JSON
func (wbo *WalletBalanceOutput) ToJSON() map[string]interface{} {
	balances := make([]map[string]interface{}, len(wbo.WalletBalanceData.Balances))
	for i, balance := range wbo.WalletBalanceData.Balances {
		balances[i] = map[string]interface{}{
			"coin_type":     balance.CoinType.String(),
			"total_balance": balance.TotalBalance.String(),
		}
	}

	return map[string]interface{}{
		"type":      wbo.Type,
		"status":    wbo.Status,
		"timestamp": wbo.Timestamp,
		"data": map[string]interface{}{
			"address_index": wbo.WalletBalanceData.AddressIndex,
			"address":       wbo.WalletBalanceData.Address,
			"balances":      balances,
		},
	}
}

// ToTable returns the wallet balance output as table rows
func (wbo *WalletBalanceOutput) ToTable() [][]string {
	rows := [][]string{
		{"Address Index", "Address"},
		{strconv.FormatUint(uint64(wbo.WalletBalanceData.AddressIndex), 10), wbo.WalletBalanceData.Address},
	}

	// Add empty row for spacing
	rows = append(rows, []string{"", ""})

	// Add balances header
	rows = append(rows, []string{"Native Assets", ""})

	// Add each balance
	for _, balance := range wbo.WalletBalanceData.Balances {
		rows = append(rows, []string{balance.CoinType.String(), balance.TotalBalance.String()})
	}

	return rows
}

// Validate validates the wallet balance output
func (wbo *WalletBalanceOutput) Validate() error {
	// Validate base output
	if err := wbo.BaseOutput.Validate(); err != nil {
		return err
	}

	// Validate wallet balance-specific fields
	if wbo.WalletBalanceData.Address == "" {
		return fmt.Errorf("address cannot be empty for wallet balance output")
	}

	// For success status, balances should be present (can be empty array)
	if wbo.Status == "success" && wbo.WalletBalanceData.Balances == nil {
		return fmt.Errorf("balances cannot be nil for successful wallet balance output")
	}

	return nil
}

// GetAddress returns the wallet address
func (wbo *WalletBalanceOutput) GetAddress() string {
	return wbo.WalletBalanceData.Address
}

// GetAddressIndex returns the address index
func (wbo *WalletBalanceOutput) GetAddressIndex() uint32 {
	return wbo.WalletBalanceData.AddressIndex
}

// GetBalances returns the balances
func (wbo *WalletBalanceOutput) GetBalances() []*iotajsonrpc.Balance {
	return wbo.WalletBalanceData.Balances
}

// GetBalanceForCoinType returns the balance for a specific coin type
func (wbo *WalletBalanceOutput) GetBalanceForCoinType(coinType string) (string, bool) {
	for _, balance := range wbo.WalletBalanceData.Balances {
		if balance.CoinType.String() == coinType {
			return balance.TotalBalance.String(), true
		}
	}
	return "0", false
}

// IsSuccess returns true if the wallet balance retrieval was successful
func (wbo *WalletBalanceOutput) IsSuccess() bool {
	return wbo.Status == "success"
}
