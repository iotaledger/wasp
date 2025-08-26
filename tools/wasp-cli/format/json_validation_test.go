package format

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWalletAddressValidation(t *testing.T) {
	validator := NewOutputValidator()

	t.Run("valid wallet address output", func(t *testing.T) {
		walletAddressOutput := NewWalletAddressSuccess(1, "iota1qp9427varyc05py79ajku89xarfgkj74tpel5egr6at6cu92mn5nku27mhf")

		err := walletAddressOutput.Validate()
		require.NoError(t, err)

		jsonData := walletAddressOutput.ToJSON()
		err = validator.ValidateJSON(jsonData)
		require.NoError(t, err)
	})

	t.Run("invalid wallet address output - empty address", func(t *testing.T) {
		walletAddressOutput := NewWalletAddressSuccess(0, "") // Empty address

		err := walletAddressOutput.Validate()
		require.Error(t, err)
		require.Contains(t, err.Error(), "address cannot be empty")
	})

	t.Run("invalid wallet address output - missing address_index", func(t *testing.T) {
		// Create malformed JSON data
		jsonData := map[string]interface{}{
			"type":      "wallet_address",
			"status":    "success",
			"timestamp": "2025-08-26T13:00:00Z",
			"data": map[string]interface{}{
				"address": "iota1qp9427varyc05py79ajku89xarfgkj74tpel5egr6at6cu92mn5nku27mhf",
				// Missing address_index
			},
		}

		err := validator.ValidateJSON(jsonData)
		require.Error(t, err)
		require.Contains(t, err.Error(), "address_index")
	})

	t.Run("error wallet address output", func(t *testing.T) {
		walletAddressOutput := NewWalletAddressError(0, "Failed to get wallet address")

		err := walletAddressOutput.Validate()
		require.NoError(t, err)

		jsonData := walletAddressOutput.ToJSON()
		err = validator.ValidateJSON(jsonData)
		require.NoError(t, err)
	})
}

func TestWalletBalanceValidation(t *testing.T) {
	validator := NewOutputValidator()

	t.Run("valid wallet balance output", func(t *testing.T) {
		walletBalanceOutput := NewWalletBalanceSuccess(1, "iota1qp9427varyc05py79ajku89xarfgkj74tpel5egr6at6cu92mn5nku27mhf", nil)

		err := walletBalanceOutput.Validate()
		require.NoError(t, err)

		jsonData := walletBalanceOutput.ToJSON()
		err = validator.ValidateJSON(jsonData)
		require.NoError(t, err)
	})

	t.Run("invalid wallet balance output - missing address_index", func(t *testing.T) {
		// Create malformed JSON data
		jsonData := map[string]interface{}{
			"type":      "wallet_balance",
			"status":    "success",
			"timestamp": "2025-08-26T13:00:00Z",
			"data": map[string]interface{}{
				"address":  "iota1qp9427varyc05py79ajku89xarfgkj74tpel5egr6at6cu92mn5nku27mhf",
				"balances": []interface{}{},
				// Missing address_index
			},
		}

		err := validator.ValidateJSON(jsonData)
		require.Error(t, err)
		require.Contains(t, err.Error(), "address_index")
	})

	t.Run("invalid wallet balance output - missing balances for success", func(t *testing.T) {
		// Create malformed JSON data
		jsonData := map[string]interface{}{
			"type":      "wallet_balance",
			"status":    "success",
			"timestamp": "2025-08-26T13:00:00Z",
			"data": map[string]interface{}{
				"address_index": 1,
				"address":       "iota1qp9427varyc05py79ajku89xarfgkj74tpel5egr6at6cu92mn5nku27mhf",
				// Missing balances
			},
		}

		err := validator.ValidateJSON(jsonData)
		require.Error(t, err)
		require.Contains(t, err.Error(), "balances")
	})

	t.Run("error wallet balance output", func(t *testing.T) {
		walletBalanceOutput := NewWalletBalanceError(0, "iota1qp9427varyc05py79ajku89xarfgkj74tpel5egr6at6cu92mn5nku27mhf", "Failed to fetch balance")

		err := walletBalanceOutput.Validate()
		require.NoError(t, err)

		jsonData := walletBalanceOutput.ToJSON()
		err = validator.ValidateJSON(jsonData)
		require.NoError(t, err)
	})
}

func TestValidationIntegration(t *testing.T) {
	t.Run("wallet address with formatter", func(t *testing.T) {
		walletAddressOutput := NewWalletAddressSuccess(0, "iota1qp9427varyc05py79ajku89xarfgkj74tpel5egr6at6cu92mn5nku27mhf")
		formatter := NewFormatter()

		// Test JSON formatting with validation
		jsonStr, err := formatter.FormatJSON(walletAddressOutput)
		require.NoError(t, err)
		require.NotEmpty(t, jsonStr)
		require.Contains(t, jsonStr, "wallet_address")
		require.Contains(t, jsonStr, "success")
	})

	t.Run("wallet balance with formatter", func(t *testing.T) {
		walletBalanceOutput := NewWalletBalanceSuccess(0, "iota1qp9427varyc05py79ajku89xarfgkj74tpel5egr6at6cu92mn5nku27mhf", nil)
		formatter := NewFormatter()

		// Test JSON formatting with validation
		jsonStr, err := formatter.FormatJSON(walletBalanceOutput)
		require.NoError(t, err)
		require.NotEmpty(t, jsonStr)
		require.Contains(t, jsonStr, "wallet_balance")
		require.Contains(t, jsonStr, "success")
	})
}
