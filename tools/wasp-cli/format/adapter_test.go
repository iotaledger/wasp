package format

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

// Mock implementations for testing
type mockCLIOutput struct {
	text string
}

func (m *mockCLIOutput) AsText() (string, error) {
	return m.text, nil
}

type mockExtendedCLIOutput struct {
	text     string
	jsonData map[string]interface{}
}

func (m *mockExtendedCLIOutput) AsText() (string, error) {
	return m.text, nil
}

func (m *mockExtendedCLIOutput) AsJSON() (string, error) {
	jsonBytes, err := json.Marshal(m.jsonData)
	return string(jsonBytes), err
}

func TestCLIOutputAdapter_BasicOutput(t *testing.T) {
	mockOutput := &mockCLIOutput{
		text: "Test output message",
	}

	adapter := NewCLIOutputAdapter(mockOutput, "test")

	t.Run("JSON output", func(t *testing.T) {
		jsonData := adapter.ToJSON()

		// Verify standard structure
		require.Equal(t, "test", jsonData["type"])
		require.Equal(t, "success", jsonData["status"])
		require.NotEmpty(t, jsonData["timestamp"])

		// Verify data structure
		data, ok := jsonData["data"].(map[string]interface{})
		require.True(t, ok)
		require.Equal(t, "Test output message", data["output"])
	})

	t.Run("Table output", func(t *testing.T) {
		tableData := adapter.ToTable()
		require.NotEmpty(t, tableData)

		// Should have header row
		require.Equal(t, "Type", tableData[0][0])
		require.Equal(t, "Output", tableData[0][1])

		// Should have data row
		require.Equal(t, "test", tableData[1][0])
		require.Equal(t, "Test output message", tableData[1][1])
	})

	t.Run("Validation", func(t *testing.T) {
		err := adapter.Validate()
		require.NoError(t, err)
	})
}

func TestCLIOutputAdapter_ExtendedOutput(t *testing.T) {
	mockOutput := &mockExtendedCLIOutput{
		text: "Extended output message",
		jsonData: map[string]interface{}{
			"field1": "value1",
			"field2": 42,
			"field3": true,
		},
	}

	adapter := NewCLIOutputAdapter(mockOutput, "extended_test")

	t.Run("JSON output with extended data", func(t *testing.T) {
		jsonData := adapter.ToJSON()

		// Verify standard structure
		require.Equal(t, "extended_test", jsonData["type"])
		require.Equal(t, "success", jsonData["status"])
		require.NotEmpty(t, jsonData["timestamp"])

		// Verify extended data is preserved
		data, ok := jsonData["data"].(map[string]interface{})
		require.True(t, ok)
		require.Equal(t, "value1", data["field1"])
		require.Equal(t, float64(42), data["field2"]) // JSON numbers are float64
		require.Equal(t, true, data["field3"])
	})

	t.Run("Table output", func(t *testing.T) {
		tableData := adapter.ToTable()
		require.NotEmpty(t, tableData)

		// Should still use text output for table format
		require.Equal(t, "extended_test", tableData[1][0])
		require.Equal(t, "Extended output message", tableData[1][1])
	})
}

func TestCLIOutputAdapter_MultilineOutput(t *testing.T) {
	mockOutput := &mockCLIOutput{
		text: "Line 1\nLine 2\nLine 3",
	}

	adapter := NewCLIOutputAdapter(mockOutput, "multiline")

	t.Run("JSON output", func(t *testing.T) {
		jsonData := adapter.ToJSON()
		data, ok := jsonData["data"].(map[string]interface{})
		require.True(t, ok)
		require.Equal(t, "Line 1\nLine 2\nLine 3", data["output"])
	})

	t.Run("Table output with multiple lines", func(t *testing.T) {
		tableData := adapter.ToTable()
		require.Len(t, tableData, 4) // Header + 3 data rows

		// Header row
		require.Equal(t, "Type", tableData[0][0])
		require.Equal(t, "Output", tableData[0][1])

		// First line with type
		require.Equal(t, "multiline", tableData[1][0])
		require.Equal(t, "Line 1", tableData[1][1])

		// Subsequent lines without type
		require.Equal(t, "", tableData[2][0])
		require.Equal(t, "Line 2", tableData[2][1])

		require.Equal(t, "", tableData[3][0])
		require.Equal(t, "Line 3", tableData[3][1])
	})
}

func TestCLIOutputAdapter_ErrorOutput(t *testing.T) {
	mockOutput := &mockCLIOutput{
		text: "Error occurred",
	}

	adapter := NewCLIOutputAdapterWithError(mockOutput, "error_test", "Something went wrong")

	t.Run("JSON output with error", func(t *testing.T) {
		jsonData := adapter.ToJSON()

		// Verify error status
		require.Equal(t, "error_test", jsonData["type"])
		require.Equal(t, "error", jsonData["status"])
		require.NotEmpty(t, jsonData["timestamp"])

		// Verify error data
		data, ok := jsonData["data"].(map[string]interface{})
		require.True(t, ok)
		require.Equal(t, "Something went wrong", data["error"])
		require.Equal(t, "Error occurred", data["output"])
	})

	t.Run("Validation", func(t *testing.T) {
		err := adapter.Validate()
		require.NoError(t, err)
	})
}

func TestCLIOutputAdapter_ValidationErrors(t *testing.T) {
	t.Run("Nil legacy output", func(t *testing.T) {
		adapter := &CLIOutputAdapter{
			BaseOutput:   NewBaseOutput("test", "success", nil),
			legacyOutput: nil,
			commandType:  "test",
		}

		err := adapter.Validate()
		require.Error(t, err)
		require.Contains(t, err.Error(), "legacy output cannot be nil")
	})
}

func TestPrintLegacyOutput(t *testing.T) {
	mockOutput := &mockCLIOutput{
		text: "Legacy output test",
	}

	// This would normally print to stdout, but we're just testing it doesn't error
	err := PrintLegacyOutput(mockOutput, "legacy_test")
	require.NoError(t, err)
}

func TestPrintLegacyError(t *testing.T) {
	mockOutput := &mockCLIOutput{
		text: "Legacy error test",
	}

	// This would normally print to stdout, but we're just testing it doesn't error
	err := PrintLegacyError(mockOutput, "legacy_error_test", "Test error message")
	require.NoError(t, err)
}

func TestCLIOutputAdapter_GetLegacyOutput(t *testing.T) {
	mockOutput := &mockCLIOutput{
		text: "Test output",
	}

	adapter := NewCLIOutputAdapter(mockOutput, "test")

	// Verify we can get the original legacy output
	legacyOutput := adapter.GetLegacyOutput()
	require.Equal(t, mockOutput, legacyOutput)

	text, err := legacyOutput.AsText()
	require.NoError(t, err)
	require.Equal(t, "Test output", text)
}

func TestCLIOutputAdapter_EmptyOutput(t *testing.T) {
	mockOutput := &mockCLIOutput{
		text: "",
	}

	adapter := NewCLIOutputAdapter(mockOutput, "empty")

	t.Run("JSON output with empty text", func(t *testing.T) {
		jsonData := adapter.ToJSON()
		data, ok := jsonData["data"].(map[string]interface{})
		require.True(t, ok)
		require.Equal(t, "", data["output"])
	})

	t.Run("Table output with empty text", func(t *testing.T) {
		tableData := adapter.ToTable()
		require.Len(t, tableData, 2) // Header + 1 empty data row

		require.Equal(t, "empty", tableData[1][0])
		require.Equal(t, "", tableData[1][1])
	})
}
