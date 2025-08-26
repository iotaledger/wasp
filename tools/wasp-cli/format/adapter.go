package format

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/iotaledger/wasp/v2/tools/wasp-cli/log"
)

// CLIOutputAdapter adapts legacy log.CLIOutput to the new CommandOutput interface
type CLIOutputAdapter struct {
	BaseOutput
	legacyOutput log.CLIOutput
	commandType  string
}

// NewCLIOutputAdapter creates a new adapter for legacy CLI output
func NewCLIOutputAdapter(legacyOutput log.CLIOutput, commandType string) *CLIOutputAdapter {
	return &CLIOutputAdapter{
		BaseOutput:   NewBaseOutput(commandType, "success", nil),
		legacyOutput: legacyOutput,
		commandType:  commandType,
	}
}

// NewCLIOutputAdapterWithError creates a new adapter for legacy CLI output with error status
func NewCLIOutputAdapterWithError(legacyOutput log.CLIOutput, commandType string, errorMsg string) *CLIOutputAdapter {
	return &CLIOutputAdapter{
		BaseOutput: NewBaseOutput(commandType, "error", map[string]interface{}{
			"error": errorMsg,
		}),
		legacyOutput: legacyOutput,
		commandType:  commandType,
	}
}

// ToJSON converts the legacy output to the standardized JSON format
func (a *CLIOutputAdapter) ToJSON() map[string]interface{} {
	var data map[string]interface{}

	// Try to get JSON output from extended CLI output
	if extended, ok := a.legacyOutput.(log.ExtendedCLIOutput); ok {
		jsonStr, err := extended.AsJSON()
		if err == nil && jsonStr != "" {
			// Parse the JSON string to extract data
			var parsed map[string]interface{}
			if err := json.Unmarshal([]byte(jsonStr), &parsed); err == nil {
				data = parsed
			}
		}
	}

	// Fallback to text output if JSON not available
	if data == nil {
		textOutput, err := a.legacyOutput.AsText()
		if err != nil {
			textOutput = fmt.Sprintf("Error getting text output: %v", err)
		}
		data = map[string]interface{}{
			"output": textOutput,
		}
	}

	// If we already have error data from constructor, merge it
	if a.Data != nil {
		if errorData, ok := a.Data.(map[string]interface{}); ok {
			for k, v := range errorData {
				data[k] = v
			}
		}
	}

	return map[string]interface{}{
		"type":      a.Type,
		"status":    a.Status,
		"timestamp": a.Timestamp,
		"data":      data,
	}
}

// ToTable converts the legacy output to table format
func (a *CLIOutputAdapter) ToTable() [][]string {
	textOutput, err := a.legacyOutput.AsText()
	if err != nil {
		textOutput = fmt.Sprintf("Error: %v", err)
	}

	// Split text into lines and create a simple table
	lines := strings.Split(strings.TrimSpace(textOutput), "\n")

	// Create table with Type and Output columns
	rows := [][]string{
		{"Type", "Output"},
	}

	// If it's a single line, put it in one row
	if len(lines) == 1 {
		rows = append(rows, []string{a.commandType, lines[0]})
	} else {
		// For multi-line output, put each line in a separate row
		for i, line := range lines {
			if i == 0 {
				rows = append(rows, []string{a.commandType, line})
			} else {
				rows = append(rows, []string{"", line})
			}
		}
	}

	return rows
}

// Validate validates the adapter output
func (a *CLIOutputAdapter) Validate() error {
	// Validate base output
	if err := a.BaseOutput.Validate(); err != nil {
		return err
	}

	// Validate that we can get text output from legacy output
	if a.legacyOutput == nil {
		return fmt.Errorf("legacy output cannot be nil")
	}

	_, err := a.legacyOutput.AsText()
	if err != nil {
		return fmt.Errorf("failed to get text output from legacy output: %w", err)
	}

	return nil
}

// GetLegacyOutput returns the wrapped legacy output for direct access if needed
func (a *CLIOutputAdapter) GetLegacyOutput() log.CLIOutput {
	return a.legacyOutput
}

// PrintLegacyOutput is a convenience function that wraps legacy output and prints it
func PrintLegacyOutput(legacyOutput log.CLIOutput, commandType string) error {
	adapter := NewCLIOutputAdapter(legacyOutput, commandType)
	return PrintOutput(adapter)
}

// PrintLegacyError is a convenience function that wraps legacy error output and prints it
func PrintLegacyError(legacyOutput log.CLIOutput, commandType string, errorMsg string) error {
	adapter := NewCLIOutputAdapterWithError(legacyOutput, commandType, errorMsg)
	return PrintOutput(adapter)
}
