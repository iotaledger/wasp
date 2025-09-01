package format

import (
	"fmt"
	"strconv"
)

// WalletAddressData represents the data structure for wallet address commands
type WalletAddressData struct {
	AddressIndex uint32 `json:"addressIndex"`
	Address      string `json:"address"`
}

// WalletAddressOutput represents the output of wallet address commands
type WalletAddressOutput struct {
	BaseOutput[WalletAddressData]
}

// NewWalletAddressOutput creates a new wallet address output
func NewWalletAddressOutput(addressIndex uint32, address string, success bool) *WalletAddressOutput {
	status := "success"
	if !success {
		status = "error"
	}

	data := WalletAddressData{
		AddressIndex: addressIndex,
		Address:      address,
	}

	return &WalletAddressOutput{
		BaseOutput: NewBaseOutput("wallet_address", status, data),
	}
}

// NewWalletAddressSuccess creates a successful wallet address output
func NewWalletAddressSuccess(addressIndex uint32, address string) *WalletAddressOutput {
	return NewWalletAddressOutput(addressIndex, address, true)
}

// ToTable returns the wallet address output as table rows
func (wao *WalletAddressOutput) ToTable() [][]string {
	return [][]string{
		{"Address Index", "Address"},
		{strconv.FormatUint(uint64(wao.Data.AddressIndex), 10), wao.Data.Address},
	}
}

// Validate validates the wallet address output
func (wao *WalletAddressOutput) Validate() error {
	// Validate base output
	if err := wao.BaseOutput.Validate(); err != nil {
		return err
	}

	// For success status, address should be present
	if wao.Status == "success" && wao.Data.Address == "" {
		return fmt.Errorf("address cannot be empty for successful wallet address output")
	}

	return nil
}

// GetAddress returns the wallet address
func (wao *WalletAddressOutput) GetAddress() string {
	return wao.Data.Address
}

// GetAddressIndex returns the address index
func (wao *WalletAddressOutput) GetAddressIndex() uint32 {
	return wao.Data.AddressIndex
}

// IsSuccess returns true if the wallet address retrieval was successful
func (wao *WalletAddressOutput) IsSuccess() bool {
	return wao.Status == "success"
}
