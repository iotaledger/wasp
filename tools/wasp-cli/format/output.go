// Package format provides utilities for formatting command output in both JSON and table formats
package format

import (
	"fmt"
	"time"
)

// CommandOutput defines the interface that all command outputs must implement
type CommandOutput interface {
	// ToJSON returns the output as a map suitable for JSON marshaling
	ToJSON() map[string]interface{}

	// ToTable returns the output as a slice of rows for table formatting
	ToTable() [][]string

	// Validate ensures the output meets the required schema
	Validate() error

	// GetType returns the command type identifier
	GetType() string

	// GetStatus returns the operation status
	GetStatus() string
}

// BaseOutput provides a standard structure for all command outputs
type BaseOutput struct {
	Type      string      `json:"type"`      // Command type (e.g., "auth", "chain", "transaction")
	Status    string      `json:"status"`    // Operation status ("success" or "error")
	Timestamp string      `json:"timestamp"` // ISO 8601 timestamp in UTC
	Data      interface{} `json:"data"`      // Command-specific data
}

// NewBaseOutput creates a new BaseOutput with the current timestamp
func NewBaseOutput(commandType, status string, data interface{}) BaseOutput {
	return BaseOutput{
		Type:      commandType,
		Status:    status,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Data:      data,
	}
}

// ToJSON returns the base output as a map suitable for JSON marshaling
func (bo BaseOutput) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"type":      bo.Type,
		"status":    bo.Status,
		"timestamp": bo.Timestamp,
		"data":      bo.Data,
	}
}

// GetType returns the command type
func (bo BaseOutput) GetType() string {
	return bo.Type
}

// GetStatus returns the operation status
func (bo BaseOutput) GetStatus() string {
	return bo.Status
}

// Validate performs basic validation on the base output structure
func (bo BaseOutput) Validate() error {
	if bo.Type == "" {
		return fmt.Errorf("command type cannot be empty")
	}

	if bo.Status != "success" && bo.Status != "error" {
		return fmt.Errorf("status must be 'success' or 'error', got: %s", bo.Status)
	}

	if bo.Timestamp == "" {
		return fmt.Errorf("timestamp cannot be empty")
	}

	// Validate timestamp format
	if _, err := time.Parse(time.RFC3339, bo.Timestamp); err != nil {
		return fmt.Errorf("invalid timestamp format: %w", err)
	}

	return nil
}

// ToTable provides a default table representation for base output
func (bo BaseOutput) ToTable() [][]string {
	return [][]string{
		{"Type", bo.Type},
		{"Status", bo.Status},
		{"Timestamp", bo.Timestamp},
	}
}

// ErrorOutput represents an error result from a command
type ErrorOutput struct {
	BaseOutput
	Error string `json:"error"`
}

// NewErrorOutput creates a new error output
func NewErrorOutput(commandType, errorMsg string) *ErrorOutput {
	return &ErrorOutput{
		BaseOutput: NewBaseOutput("error", "error", map[string]interface{}{
			"error": errorMsg,
		}),
		Error: errorMsg,
	}
}

// ToJSON returns the error output as JSON
func (eo *ErrorOutput) ToJSON() map[string]interface{} {
	result := eo.BaseOutput.ToJSON()
	result["data"] = map[string]interface{}{
		"error": eo.Error,
	}
	return result
}

// ToTable returns the error output as a table
func (eo *ErrorOutput) ToTable() [][]string {
	return [][]string{
		{"Type", eo.Type},
		{"Status", eo.Status},
		{"Error", eo.Error},
		{"Timestamp", eo.Timestamp},
	}
}

// Validate validates the error output
func (eo *ErrorOutput) Validate() error {
	if err := eo.BaseOutput.Validate(); err != nil {
		return err
	}

	if eo.Error == "" {
		return fmt.Errorf("error message cannot be empty for error output")
	}

	return nil
}
