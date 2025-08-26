package format

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/iotaledger/wasp/v2/tools/wasp-cli/log"
)

func TestGenericFormatter(t *testing.T) {
	formatter := NewFormatter()

	t.Run("auth output json format", func(t *testing.T) {
		// Enable JSON flag
		log.JSONFlag = true
		defer func() { log.JSONFlag = false }()

		authOutput := NewAuthSuccess("0", "wasp")

		result, err := formatter.Format(authOutput)
		if err != nil {
			t.Fatalf("Format failed: %v", err)
		}

		// Parse JSON to verify structure
		var jsonData map[string]interface{}
		if err := json.Unmarshal([]byte(result), &jsonData); err != nil {
			t.Fatalf("Failed to parse JSON: %v", err)
		}

		// Verify required fields
		if jsonData["type"] != "auth" {
			t.Errorf("Expected type 'auth', got: %v", jsonData["type"])
		}
		if jsonData["status"] != "success" {
			t.Errorf("Expected status 'success', got: %v", jsonData["status"])
		}
		if jsonData["timestamp"] == "" {
			t.Error("Expected non-empty timestamp")
		}

		// Verify data structure
		data, ok := jsonData["data"].(map[string]interface{})
		if !ok {
			t.Fatal("Expected data to be an object")
		}
		if data["node"] != "0" {
			t.Errorf("Expected node '0', got: %v", data["node"])
		}
		if data["username"] != "wasp" {
			t.Errorf("Expected username 'wasp', got: %v", data["username"])
		}
	})

	t.Run("auth output table format", func(t *testing.T) {
		// Disable JSON flag
		log.JSONFlag = false

		authOutput := NewAuthSuccess("0", "wasp")

		result, err := formatter.Format(authOutput)
		if err != nil {
			t.Fatalf("Format failed: %v", err)
		}

		// Verify table format
		if !strings.Contains(result, "|") {
			t.Error("Expected table format with separators")
		}
		if !strings.Contains(result, "success") {
			t.Error("Expected 'success' in table output")
		}
		if !strings.Contains(result, "wasp") {
			t.Error("Expected 'wasp' in table output")
		}
	})

	t.Run("error output", func(t *testing.T) {
		log.JSONFlag = true
		defer func() { log.JSONFlag = false }()

		errorOutput := NewErrorOutput("auth", "Invalid credentials")

		result, err := formatter.Format(errorOutput)
		if err != nil {
			t.Fatalf("Format failed: %v", err)
		}

		var jsonData map[string]interface{}
		if err := json.Unmarshal([]byte(result), &jsonData); err != nil {
			t.Fatalf("Failed to parse JSON: %v", err)
		}

		if jsonData["status"] != "error" {
			t.Errorf("Expected status 'error', got: %v", jsonData["status"])
		}

		data, ok := jsonData["data"].(map[string]interface{})
		if !ok {
			t.Fatal("Expected data to be an object")
		}
		if data["error"] != "Invalid credentials" {
			t.Errorf("Expected error message 'Invalid credentials', got: %v", data["error"])
		}
	})
}

func TestOutputValidation(t *testing.T) {
	validator := NewOutputValidator()

	t.Run("valid auth output", func(t *testing.T) {
		authOutput := NewAuthSuccess("0", "wasp")

		if err := authOutput.Validate(); err != nil {
			t.Errorf("Valid auth output failed validation: %v", err)
		}

		jsonData := authOutput.ToJSON()
		if err := validator.ValidateJSON(jsonData); err != nil {
			t.Errorf("Valid JSON failed validation: %v", err)
		}
	})

	t.Run("invalid auth output - empty node", func(t *testing.T) {
		authOutput := NewAuthSuccess("", "wasp") // Empty node

		if err := authOutput.Validate(); err == nil {
			t.Error("Expected validation to fail for empty node")
		}
	})

	t.Run("invalid auth output - empty username", func(t *testing.T) {
		authOutput := NewAuthSuccess("0", "") // Empty username

		if err := authOutput.Validate(); err == nil {
			t.Error("Expected validation to fail for empty username")
		}
	})

	t.Run("invalid status", func(t *testing.T) {
		// Create output with invalid status
		baseOutput := BaseOutput{
			Type:      "auth",
			Status:    "invalid", // Invalid status
			Timestamp: "2025-08-19T13:50:50Z",
			Data:      map[string]interface{}{},
		}

		if err := baseOutput.Validate(); err == nil {
			t.Error("Expected validation to fail for invalid status")
		}
	})

	t.Run("invalid timestamp", func(t *testing.T) {
		// Create output with invalid timestamp
		baseOutput := BaseOutput{
			Type:      "auth",
			Status:    "success",
			Timestamp: "invalid-timestamp", // Invalid timestamp
			Data:      map[string]interface{}{},
		}

		if err := baseOutput.Validate(); err == nil {
			t.Error("Expected validation to fail for invalid timestamp")
		}
	})
}

func TestConvenienceFunctions(t *testing.T) {
	t.Run("global format function", func(t *testing.T) {
		log.JSONFlag = true
		defer func() { log.JSONFlag = false }()

		authOutput := NewAuthSuccess("0", "wasp")

		result, err := Format(authOutput)
		if err != nil {
			t.Fatalf("Global Format failed: %v", err)
		}

		if !strings.Contains(result, `"type":"auth"`) {
			t.Error("Expected JSON output with auth type")
		}
	})

	t.Run("print error function", func(t *testing.T) {
		log.JSONFlag = true
		defer func() { log.JSONFlag = false }()

		// This would normally print to stdout, but we can't easily capture that in this test
		// Just verify it doesn't panic
		err := PrintError("auth", "Test error message")
		if err != nil {
			t.Errorf("PrintError failed: %v", err)
		}
	})
}

func TestTableFormatting(t *testing.T) {
	formatter := NewFormatter()

	t.Run("table with different column widths", func(t *testing.T) {
		log.JSONFlag = false

		// Create auth output with longer values to test column width calculation
		authOutput := NewAuthError("node-with-long-name", "very-long-username", "This is a very long error message")

		result, err := formatter.Format(authOutput)
		if err != nil {
			t.Fatalf("Format failed: %v", err)
		}

		lines := strings.Split(result, "\n")
		if len(lines) < 3 {
			t.Error("Expected at least 3 lines in table output")
		}

		// Check that all lines have consistent formatting
		for i, line := range lines {
			if line == "" {
				continue // Skip empty lines
			}
			if !strings.HasPrefix(line, "|") {
				t.Errorf("Line %d should start with '|': %s", i, line)
			}
			if !strings.HasSuffix(strings.TrimSpace(line), "|") {
				t.Errorf("Line %d should end with '|': %s", i, line)
			}
		}
	})
}

func TestCustomValidationRules(t *testing.T) {
	validator := NewOutputValidator()

	t.Run("add custom validation rule", func(t *testing.T) {
		// Add a custom rule that requires a specific field in auth data
		customRule := func(data map[string]interface{}) error {
			dataField, ok := data["data"].(map[string]interface{})
			if !ok {
				return nil // Skip if data is not an object
			}

			if _, exists := dataField["custom_field"]; !exists {
				return fmt.Errorf("missing required custom_field")
			}
			return nil
		}

		err := validator.AddValidationRule("auth", customRule)
		if err != nil {
			t.Fatalf("Failed to add validation rule: %v", err)
		}

		// Test that validation now fails without the custom field
		authOutput := NewAuthSuccess("0", "wasp")
		jsonData := authOutput.ToJSON()

		if err := validator.ValidateJSON(jsonData); err == nil {
			t.Error("Expected validation to fail due to missing custom_field")
		}
	})
}
