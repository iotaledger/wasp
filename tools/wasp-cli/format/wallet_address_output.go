package format

import (
	"fmt"
	"strconv"
)

// WalletAddressData represents the data structure for wallet address commands
type WalletAddressData struct {
	AddressIndex uint32 `json:"address_index"`
	Address      string `json:"address"`
}

// WalletAddressOutput represents the output of wallet address commands
type WalletAddressOutput struct {
	BaseOutput
	WalletAddressData WalletAddressData `json:"-"` // Embedded for easy access, but not serialized directly
}

// NewWalletAddressOutput creates a new wallet address output
func NewWalletAddressOutput(addressIndex uint32, address string, success bool) *WalletAddressOutput {
	status := "success"
	if !success {
		status = "error"
	}

	walletAddressData := WalletAddressData{
		AddressIndex: addressIndex,
		Address:      address,
	}

	return &WalletAddressOutput{
		BaseOutput:        NewBaseOutput("wallet_address", status, walletAddressData),
		WalletAddressData: walletAddressData,
	}
}

// NewWalletAddressSuccess creates a successful wallet address output
func NewWalletAddressSuccess(addressIndex uint32, address string) *WalletAddressOutput {
	return NewWalletAddressOutput(addressIndex, address, true)
}

// NewWalletAddressError creates an error wallet address output
func NewWalletAddressError(addressIndex uint32, errorMsg string) *WalletAddressOutput {
	return &WalletAddressOutput{
		BaseOutput: NewBaseOutput("wallet_address", "error", map[string]interface{}{
			"address_index": addressIndex,
			"error":         errorMsg,
		}),
		WalletAddressData: WalletAddressData{
			AddressIndex: addressIndex,
			Address:      "",
		},
	}
}

// ToJSON returns the wallet address output as JSON
func (wao *WalletAddressOutput) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"type":      wao.Type,
		"status":    wao.Status,
		"timestamp": wao.Timestamp,
		"data": map[string]interface{}{
			"address_index": wao.WalletAddressData.AddressIndex,
			"address":       wao.WalletAddressData.Address,
		},
	}
}

// ToTable returns the wallet address output as table rows
func (wao *WalletAddressOutput) ToTable() [][]string {
	return [][]string{
		{"Address Index", "Address"},
		{strconv.FormatUint(uint64(wao.WalletAddressData.AddressIndex), 10), wao.WalletAddressData.Address},
	}
}

// Validate validates the wallet address output
func (wao *WalletAddressOutput) Validate() error {
	// Validate base output
	if err := wao.BaseOutput.Validate(); err != nil {
		return err
	}

	// For success status, address should be present
	if wao.Status == "success" && wao.WalletAddressData.Address == "" {
		return fmt.Errorf("address cannot be empty for successful wallet address output")
	}

	return nil
}

// GetAddress returns the wallet address
func (wao *WalletAddressOutput) GetAddress() string {
	return wao.WalletAddressData.Address
}

// GetAddressIndex returns the address index
func (wao *WalletAddressOutput) GetAddressIndex() uint32 {
	return wao.WalletAddressData.AddressIndex
}

// IsSuccess returns true if the wallet address retrieval was successful
func (wao *WalletAddressOutput) IsSuccess() bool {
	return wao.Status == "success"
}
