package format

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/iotaledger/wasp/v2/tools/wasp-cli/log"
)

// Formatter provides a unified interface for formatting command outputs
type Formatter struct {
	validator *OutputValidator
}

// NewFormatter creates a new formatter with validation
func NewFormatter() *Formatter {
	return &Formatter{
		validator: NewOutputValidator(),
	}
}

// Format formats the command output based on the global JSON flag
func (f *Formatter) Format(output CommandOutput) (string, error) {
	// Validate the output first
	if err := output.Validate(); err != nil {
		return "", fmt.Errorf("output validation failed: %w", err)
	}

	if log.JSONFlag {
		return f.FormatJSON(output)
	}
	return f.FormatTable(output), nil
}

// FormatJSON formats the output as JSON
func (f *Formatter) FormatJSON(output CommandOutput) (string, error) {
	jsonData := output.ToJSON()

	// Validate the JSON structure
	if err := f.validator.ValidateJSON(jsonData); err != nil {
		return "", fmt.Errorf("JSON validation failed: %w", err)
	}

	// Standardize the output format
	standardized := f.standardizeOutput(jsonData)

	// Marshal to JSON with proper formatting
	jsonBytes, err := json.Marshal(standardized)
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return string(jsonBytes), nil
}

// FormatTable formats the output as a table
func (f *Formatter) FormatTable(output CommandOutput) string {
	rows := output.ToTable()
	if len(rows) == 0 {
		return ""
	}

	// Calculate column widths
	maxWidths := make([]int, len(rows[0]))
	for _, row := range rows {
		for i, cell := range row {
			if len(cell) > maxWidths[i] {
				maxWidths[i] = len(cell)
			}
		}
	}

	var result strings.Builder

	// Format each row
	for i, row := range rows {
		result.WriteString("| ")
		for j, cell := range row {
			result.WriteString(fmt.Sprintf("%-*s | ", maxWidths[j], cell))
		}
		result.WriteString("\n")

		// Add separator after first row (headers)
		if i == 0 && len(rows) > 1 {
			result.WriteString("|")
			for j := range row {
				result.WriteString(strings.Repeat("-", maxWidths[j]+2) + "|")
			}
			result.WriteString("\n")
		}
	}

	return result.String()
}

// standardizeOutput ensures the JSON output follows the standard format
func (f *Formatter) standardizeOutput(data map[string]interface{}) map[string]interface{} {
	// Ensure required fields are present
	standardized := make(map[string]interface{})

	// Copy all fields
	for k, v := range data {
		standardized[k] = v
	}

	// Ensure type field exists
	if _, exists := standardized["type"]; !exists {
		standardized["type"] = "unknown"
	}

	// Ensure status field exists
	if _, exists := standardized["status"]; !exists {
		standardized["status"] = "unknown"
	}

	// Ensure timestamp field exists
	if _, exists := standardized["timestamp"]; !exists {
		standardized["timestamp"] = ""
	}

	// Ensure data field exists
	if _, exists := standardized["data"]; !exists {
		standardized["data"] = map[string]interface{}{}
	}

	return standardized
}

// PrintOutput is a convenience function that formats and prints the output
func (f *Formatter) PrintOutput(output CommandOutput) error {
	formatted, err := f.Format(output)
	if err != nil {
		return err
	}

	fmt.Print(formatted)
	return nil
}

// PrintError is a convenience function for printing error outputs
func (f *Formatter) PrintError(commandType, errorMsg string) error {
	errorOutput := NewErrorOutput(commandType, errorMsg)
	return f.PrintOutput(errorOutput)
}

// Global formatter instance for convenience
var defaultFormatter = NewFormatter()

// Format formats output using the default formatter
func Format(output CommandOutput) (string, error) {
	return defaultFormatter.Format(output)
}

// PrintOutput prints output using the default formatter
func PrintOutput(output CommandOutput) error {
	return defaultFormatter.PrintOutput(output)
}

// PrintError prints an error using the default formatter
func PrintError(commandType, errorMsg string) error {
	return defaultFormatter.PrintError(commandType, errorMsg)
}
