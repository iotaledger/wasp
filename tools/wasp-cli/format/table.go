// Package format provides utilities for formatting output using glazed
package format

import (
	"context"

	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
)

// TableFormatter provides a simple interface for formatting data as tables
type TableFormatter struct {
	gp *middlewares.TableProcessor
}

// NewTableFormatter creates a new table formatter with default settings
func NewTableFormatter() (*TableFormatter, error) {
	// Create table processor with default settings
	gp := middlewares.NewTableProcessor()

	return &TableFormatter{gp: gp}, nil
}

// AddRow adds a row of data to the table
func (tf *TableFormatter) AddRow(data map[string]interface{}) error {
	row := types.NewRowFromMap(data)
	return tf.gp.AddRow(context.Background(), row)
}

// Render outputs the formatted table to stdout
func (tf *TableFormatter) Render() error {
	return tf.gp.Close(context.Background())
}

// AuthResult represents the result of an authentication operation
type AuthResult struct {
	Status   string `json:"status"`
	Node     string `json:"node"`
	Username string `json:"username"`
	Message  string `json:"message,omitempty"`
}

// FormatAuthResult formats an authentication result using the new glazed formatter
func FormatAuthResult(result AuthResult) error {
	formatter := NewGlazedFormatter()
	return formatter.FormatAuthResult(result.Status, result.Node, result.Username, result.Message)
}

// FormatAuthError formats an error message using the new glazed formatter
func FormatAuthError(operation, node, username, errorMsg string) error {
	formatter := NewGlazedFormatter()
	return formatter.FormatAuthResult("error", node, username, errorMsg)
}

// FormatAuthSuccess formats a success message using the new glazed formatter
func FormatAuthSuccess(node, username string) error {
	formatter := NewGlazedFormatter()
	return formatter.FormatAuthResult("success", node, username, "Authentication successful")
}
