package format

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWalletAddressOutput_Success(t *testing.T) {
	output := NewWalletAddressSuccess(1, "iota1qp9427varyc05py79ajku89xarfgkj74tpel5egr6at6cu92mn5nku27mhf")

	t.Run("JSON output", func(t *testing.T) {
		jsonData := output.ToJSON()

		// Verify top-level structure
		require.Equal(t, "wallet_address", jsonData["type"])
		require.Equal(t, "success", jsonData["status"])
		require.NotEmpty(t, jsonData["timestamp"])

		// Verify data structure
		data, ok := jsonData["data"].(map[string]interface{})
		require.True(t, ok)
		require.Equal(t, uint32(1), data["address_index"])
		require.Equal(t, "iota1qp9427varyc05py79ajku89xarfgkj74tpel5egr6at6cu92mn5nku27mhf", data["address"])

		// Verify JSON can be marshaled
		jsonBytes, err := json.Marshal(jsonData)
		require.NoError(t, err)
		require.NotEmpty(t, jsonBytes)
	})

	t.Run("Table output", func(t *testing.T) {
		tableData := output.ToTable()
		require.NotEmpty(t, tableData)

		// Should have header row and address row
		require.Len(t, tableData, 2)

		// Check header
		require.Equal(t, "Address Index", tableData[0][0])
		require.Equal(t, "Address", tableData[0][1])

		// Check address data
		require.Equal(t, "1", tableData[1][0])
		require.Equal(t, "iota1qp9427varyc05py79ajku89xarfgkj74tpel5egr6at6cu92mn5nku27mhf", tableData[1][1])
	})

	t.Run("Validation", func(t *testing.T) {
		err := output.Validate()
		require.NoError(t, err)
	})

	t.Run("Getter methods", func(t *testing.T) {
		require.Equal(t, "iota1qp9427varyc05py79ajku89xarfgkj74tpel5egr6at6cu92mn5nku27mhf", output.GetAddress())
		require.Equal(t, uint32(1), output.GetAddressIndex())
		require.True(t, output.IsSuccess())
	})
}

func TestWalletAddressOutput_Error(t *testing.T) {
	output := NewWalletAddressError(0, "Failed to get wallet address")

	t.Run("JSON output", func(t *testing.T) {
		jsonData := output.ToJSON()

		// Verify top-level structure
		require.Equal(t, "wallet_address", jsonData["type"])
		require.Equal(t, "error", jsonData["status"])
		require.NotEmpty(t, jsonData["timestamp"])

		// Verify data structure contains error info
		data, ok := jsonData["data"].(map[string]interface{})
		require.True(t, ok)
		require.Equal(t, uint32(0), data["address_index"])
		require.Equal(t, "Failed to get wallet address", data["error"])
	})

	t.Run("Validation", func(t *testing.T) {
		err := output.Validate()
		require.NoError(t, err)
	})

	t.Run("Getter methods", func(t *testing.T) {
		require.Equal(t, "", output.GetAddress())
		require.Equal(t, uint32(0), output.GetAddressIndex())
		require.False(t, output.IsSuccess())
	})
}

func TestWalletAddressOutput_Validation_Errors(t *testing.T) {
	t.Run("Empty address for success", func(t *testing.T) {
		output := NewWalletAddressSuccess(0, "")
		err := output.Validate()
		require.Error(t, err)
		require.Contains(t, err.Error(), "address cannot be empty")
	})

	t.Run("Invalid base output", func(t *testing.T) {
		output := &WalletAddressOutput{
			BaseOutput: NewBaseOutput("", "invalid_status", nil),
			WalletAddressData: WalletAddressData{
				AddressIndex: 0,
				Address:      "test_address",
			},
		}
		err := output.Validate()
		require.Error(t, err)
	})
}

func TestWalletAddressOutput_ZeroIndex(t *testing.T) {
	// Test with address index 0 (which is valid)
	output := NewWalletAddressSuccess(0, "iota1qp9427varyc05py79ajku89xarfgkj74tpel5egr6at6cu92mn5nku27mhf")

	t.Run("JSON output", func(t *testing.T) {
		jsonData := output.ToJSON()
		data, ok := jsonData["data"].(map[string]interface{})
		require.True(t, ok)
		require.Equal(t, uint32(0), data["address_index"])
		require.Equal(t, "iota1qp9427varyc05py79ajku89xarfgkj74tpel5egr6at6cu92mn5nku27mhf", data["address"])
	})

	t.Run("Table output", func(t *testing.T) {
		tableData := output.ToTable()
		require.Equal(t, "0", tableData[1][0])
		require.Equal(t, "iota1qp9427varyc05py79ajku89xarfgkj74tpel5egr6at6cu92mn5nku27mhf", tableData[1][1])
	})

	t.Run("Validation", func(t *testing.T) {
		err := output.Validate()
		require.NoError(t, err)
	})
}

func TestWalletAddressOutput_CommandOutputInterface(t *testing.T) {
	// Test that WalletAddressOutput implements CommandOutput interface
	var output CommandOutput = NewWalletAddressSuccess(1, "iota1qp9427varyc05py79ajku89xarfgkj74tpel5egr6at6cu92mn5nku27mhf")

	// Test interface methods
	require.Equal(t, "wallet_address", output.GetType())
	require.Equal(t, "success", output.GetStatus())
	require.NotNil(t, output.ToJSON())
	require.NotNil(t, output.ToTable())
	require.NoError(t, output.Validate())
}
