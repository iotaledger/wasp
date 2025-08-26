package format

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/iotaledger/wasp/v2/tools/wasp-cli/log"
)

func TestFormatAuthResult(t *testing.T) {
	// Test data
	result := AuthResult{
		Status:   "SUCCESS",
		Node:     "0",
		Username: "wasp",
		Message:  "Authentication successful",
	}

	t.Run("table format", func(t *testing.T) {
		// Ensure JSON flag is false
		log.JSONFlag = false

		// Capture stdout
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := FormatAuthResult(result)
		if err != nil {
			t.Fatalf("FormatAuthResult failed: %v", err)
		}

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		// Check that output contains expected table elements
		if !strings.Contains(output, "SUCCESS") {
			t.Errorf("Expected table output to contain 'SUCCESS', got: %s", output)
		}
		if !strings.Contains(output, "wasp") {
			t.Errorf("Expected table output to contain 'wasp', got: %s", output)
		}
		if !strings.Contains(output, "|") {
			t.Errorf("Expected table output to contain table separators '|', got: %s", output)
		}
	})

	t.Run("json format", func(t *testing.T) {
		// Enable JSON flag
		log.JSONFlag = true
		defer func() { log.JSONFlag = false }() // Reset after test

		// Capture stdout
		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := FormatAuthResult(result)
		if err != nil {
			t.Fatalf("FormatAuthResult failed: %v", err)
		}

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		buf.ReadFrom(r)
		output := buf.String()

		// Check that output is valid JSON
		if !strings.Contains(output, `"status"`) {
			t.Errorf("Expected JSON output to contain '\"status\"', got: %s", output)
		}
		if !strings.Contains(output, `"SUCCESS"`) {
			t.Errorf("Expected JSON output to contain '\"SUCCESS\"', got: %s", output)
		}
		if !strings.Contains(output, `"node"`) {
			t.Errorf("Expected JSON output to contain '\"node\"', got: %s", output)
		}
		if !strings.Contains(output, `"username"`) {
			t.Errorf("Expected JSON output to contain '\"username\"', got: %s", output)
		}
		if !strings.Contains(output, `"wasp"`) {
			t.Errorf("Expected JSON output to contain '\"wasp\"', got: %s", output)
		}

		// Should not contain table separators
		if strings.Contains(output, "|") {
			t.Errorf("Expected JSON output to not contain table separators '|', got: %s", output)
		}
	})
}
