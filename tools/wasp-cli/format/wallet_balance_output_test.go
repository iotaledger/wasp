package format

import (
	"encoding/json"
	"testing"

	"github.com/iotaledger/wasp/v2/clients/iota-go/iotajsonrpc"
	"github.com/stretchr/testify/require"
)

func TestWalletBalanceOutput_Success(t *testing.T) {
	// Create test balance data
	balances := []*iotajsonrpc.Balance{
		{
			CoinType:     iotajsonrpc.MustCoinTypeFromString("0x0000000000000000000000000000000000000000000000000000000000000002::iota::IOTA"),
			TotalBalance: iotajsonrpc.NewBigInt(1000),
		},
		{
			CoinType:     iotajsonrpc.MustCoinTypeFromString("0x1234567890abcdef1234567890abcdef12345678::test::TOKEN"),
			TotalBalance: iotajsonrpc.NewBigInt(500),
		},
	}

	output := NewWalletBalanceSuccess(1, "iota1qp9427varyc05py79ajku89xarfgkj74tpel5egr6at6cu92mn5nku27mhf", balances)

	t.Run("JSON output", func(t *testing.T) {
		jsonData := output.ToJSON()

		// Verify top-level structure
		require.Equal(t, "wallet_balance", jsonData["type"])
		require.Equal(t, "success", jsonData["status"])
		require.NotEmpty(t, jsonData["timestamp"])

		// Verify data structure
		data, ok := jsonData["data"].(map[string]interface{})
		require.True(t, ok)
		require.Equal(t, uint32(1), data["address_index"])
		require.Equal(t, "iota1qp9427varyc05py79ajku89xarfgkj74tpel5egr6at6cu92mn5nku27mhf", data["address"])

		// Verify balances
		balancesData, ok := data["balances"].([]map[string]interface{})
		require.True(t, ok)
		require.Len(t, balancesData, 2)

		// Check first balance
		require.Equal(t, "0x0000000000000000000000000000000000000000000000000000000000000002::iota::IOTA", balancesData[0]["coin_type"])
		require.Equal(t, "1000", balancesData[0]["total_balance"])

		// Check second balance
		require.Equal(t, "0x1234567890abcdef1234567890abcdef12345678::test::TOKEN", balancesData[1]["coin_type"])
		require.Equal(t, "500", balancesData[1]["total_balance"])

		// Verify JSON can be marshaled
		jsonBytes, err := json.Marshal(jsonData)
		require.NoError(t, err)
		require.NotEmpty(t, jsonBytes)
	})

	t.Run("Table output", func(t *testing.T) {
		tableData := output.ToTable()
		require.NotEmpty(t, tableData)

		// Should have header row, address row, empty row, assets header, and balance rows
		require.GreaterOrEqual(t, len(tableData), 5)

		// Check header
		require.Equal(t, "Address Index", tableData[0][0])
		require.Equal(t, "Address", tableData[0][1])

		// Check address data
		require.Equal(t, "1", tableData[1][0])
		require.Equal(t, "iota1qp9427varyc05py79ajku89xarfgkj74tpel5egr6at6cu92mn5nku27mhf", tableData[1][1])

		// Check assets header
		require.Equal(t, "Native Assets", tableData[3][0])

		// Check balance entries
		require.Contains(t, tableData[4][0], "::iota::IOTA")
		require.Equal(t, "1000", tableData[4][1])
	})

	t.Run("Validation", func(t *testing.T) {
		err := output.Validate()
		require.NoError(t, err)
	})

	t.Run("Getter methods", func(t *testing.T) {
		require.Equal(t, "iota1qp9427varyc05py79ajku89xarfgkj74tpel5egr6at6cu92mn5nku27mhf", output.GetAddress())
		require.Equal(t, uint32(1), output.GetAddressIndex())
		require.Equal(t, balances, output.GetBalances())
		require.True(t, output.IsSuccess())

		// Test GetBalanceForCoinType
		balance, found := output.GetBalanceForCoinType("0x0000000000000000000000000000000000000000000000000000000000000002::iota::IOTA")
		require.True(t, found)
		require.Equal(t, "1000", balance)

		balance, found = output.GetBalanceForCoinType("nonexistent")
		require.False(t, found)
		require.Equal(t, "0", balance)
	})
}

func TestWalletBalanceOutput_Error(t *testing.T) {
	output := NewWalletBalanceError(0, "iota1qp9427varyc05py79ajku89xarfgkj74tpel5egr6at6cu92mn5nku27mhf", "Failed to fetch balance")

	t.Run("JSON output", func(t *testing.T) {
		jsonData := output.ToJSON()

		// Verify top-level structure
		require.Equal(t, "wallet_balance", jsonData["type"])
		require.Equal(t, "error", jsonData["status"])
		require.NotEmpty(t, jsonData["timestamp"])

		// Verify data structure contains error info
		data, ok := jsonData["data"].(map[string]interface{})
		require.True(t, ok)
		require.Equal(t, uint32(0), data["address_index"])
		require.Equal(t, "iota1qp9427varyc05py79ajku89xarfgkj74tpel5egr6at6cu92mn5nku27mhf", data["address"])
		require.Equal(t, "Failed to fetch balance", data["error"])
	})

	t.Run("Validation", func(t *testing.T) {
		err := output.Validate()
		require.NoError(t, err)
	})

	t.Run("Getter methods", func(t *testing.T) {
		require.Equal(t, "iota1qp9427varyc05py79ajku89xarfgkj74tpel5egr6at6cu92mn5nku27mhf", output.GetAddress())
		require.Equal(t, uint32(0), output.GetAddressIndex())
		require.Empty(t, output.GetBalances())
		require.False(t, output.IsSuccess())
	})
}

func TestWalletBalanceOutput_Validation_Errors(t *testing.T) {
	t.Run("Empty address", func(t *testing.T) {
		output := NewWalletBalanceSuccess(0, "", []*iotajsonrpc.Balance{})
		err := output.Validate()
		require.Error(t, err)
		require.Contains(t, err.Error(), "address cannot be empty")
	})

	t.Run("Nil balances for success", func(t *testing.T) {
		output := &WalletBalanceOutput{
			BaseOutput: NewBaseOutput("wallet_balance", "success", nil),
			WalletBalanceData: WalletBalanceData{
				AddressIndex: 0,
				Address:      "iota1qp9427varyc05py79ajku89xarfgkj74tpel5egr6at6cu92mn5nku27mhf",
				Balances:     nil,
			},
		}
		err := output.Validate()
		require.Error(t, err)
		require.Contains(t, err.Error(), "balances cannot be nil")
	})
}

func TestWalletBalanceOutput_EmptyBalances(t *testing.T) {
	// Test with empty but not nil balances array
	output := NewWalletBalanceSuccess(0, "iota1qp9427varyc05py79ajku89xarfgkj74tpel5egr6at6cu92mn5nku27mhf", []*iotajsonrpc.Balance{})

	t.Run("JSON output", func(t *testing.T) {
		jsonData := output.ToJSON()
		data, ok := jsonData["data"].(map[string]interface{})
		require.True(t, ok)

		balances, ok := data["balances"].([]map[string]interface{})
		require.True(t, ok)
		require.Empty(t, balances)
	})

	t.Run("Table output", func(t *testing.T) {
		tableData := output.ToTable()
		require.NotEmpty(t, tableData)
		// Should still have header and address info, just no balance entries
		require.GreaterOrEqual(t, len(tableData), 4)
	})

	t.Run("Validation", func(t *testing.T) {
		err := output.Validate()
		require.NoError(t, err)
	})
}
