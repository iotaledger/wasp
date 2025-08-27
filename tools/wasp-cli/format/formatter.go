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
	// Use Go's built-in JSON marshaling directly
	jsonBytes, err := json.Marshal(output)
	if err != nil {
		return "", fmt.Errorf("JSON marshaling failed: %w", err)
	}

	// Convert to map for validation if needed
	var jsonData map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &jsonData); err != nil {
		return "", fmt.Errorf("JSON unmarshaling for validation failed: %w", err)
	}

	// Validate the JSON structure
	if err := f.validator.ValidateJSON(jsonData); err != nil {
		return "", fmt.Errorf("JSON validation failed: %w", err)
	}

	// Return the properly marshaled JSON
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

// PrintOutput prints output using the default formatter
func PrintOutput(output CommandOutput) error {
	return defaultFormatter.PrintOutput(output)
}
