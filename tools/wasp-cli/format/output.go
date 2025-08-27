// Package format provides utilities for formatting command output in both JSON and table formats
package format

import (
	"fmt"
	"time"
)

// CommandOutput defines the interface that all command outputs must implement
type CommandOutput interface {
	// ToTable returns the output as a slice of rows for table formatting
	ToTable() [][]string

	// Validate ensures the output meets the required schema
	Validate() error

	// GetType returns the command type identifier
	GetType() string

	// GetStatus returns the operation status
	GetStatus() string
}

// BaseOutput provides a standard structure for all command outputs using generics
type BaseOutput[T any] struct {
	Type      string    `json:"type"`      // Command type (e.g., "auth", "chain", "transaction")
	Status    string    `json:"status"`    // Operation status ("success" or "error")
	Timestamp time.Time `json:"timestamp"` // ISO 8601 timestamp in UTC
	Data      T         `json:"data"`      // Command-specific data with proper type safety
}

// NewBaseOutput creates a new BaseOutput with the current timestamp
func NewBaseOutput[T any](commandType, status string, data T) BaseOutput[T] {
	return BaseOutput[T]{
		Type:      commandType,
		Status:    status,
		Timestamp: time.Now().UTC(),
		Data:      data,
	}
}

// GetType returns the command type
func (bo BaseOutput[T]) GetType() string {
	return bo.Type
}

// GetStatus returns the operation status
func (bo BaseOutput[T]) GetStatus() string {
	return bo.Status
}

// GetData returns the typed data
func (bo BaseOutput[T]) GetData() T {
	return bo.Data
}

// Validate performs basic validation on the base output structure
func (bo BaseOutput[T]) Validate() error {
	if bo.Type == "" {
		return fmt.Errorf("command type cannot be empty")
	}

	if bo.Status != "success" && bo.Status != "error" {
		return fmt.Errorf("status must be 'success' or 'error', got: %s", bo.Status)
	}

	if bo.Timestamp.IsZero() {
		return fmt.Errorf("timestamp cannot be zero")
	}

	return nil
}

// ToTable provides a default table representation for base output
func (bo BaseOutput[T]) ToTable() [][]string {
	return [][]string{
		{"Type", bo.Type},
		{"Status", bo.Status},
		{"Timestamp", bo.Timestamp.Format(time.RFC3339)},
	}
}

// ErrorData represents the data structure for error outputs
type ErrorData struct {
	Error string `json:"error"`
}

// ErrorOutput represents an error result from a command
type ErrorOutput struct {
	BaseOutput[ErrorData]
}

// NewErrorOutput creates a new error output
func NewErrorOutput(commandType, errorMsg string) *ErrorOutput {
	data := ErrorData{
		Error: errorMsg,
	}

	return &ErrorOutput{
		BaseOutput: NewBaseOutput(commandType, "error", data),
	}
}

// ToTable returns the error output as a table
func (eo *ErrorOutput) ToTable() [][]string {
	return [][]string{
		{"Type", eo.Type},
		{"Status", eo.Status},
		{"Error", eo.Data.Error},
		{"Timestamp", eo.Timestamp.Format(time.RFC3339)},
	}
}

// Validate validates the error output
func (eo *ErrorOutput) Validate() error {
	if err := eo.BaseOutput.Validate(); err != nil {
		return err
	}

	if eo.Data.Error == "" {
		return fmt.Errorf("error message cannot be empty for error output")
	}

	return nil
}

// GetError returns the error message
func (eo *ErrorOutput) GetError() string {
	return eo.Data.Error
}
