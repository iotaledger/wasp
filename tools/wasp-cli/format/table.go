// Package format provides utilities for formatting output using glazed
package format

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/log"
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

// FormatAuthResult formats an authentication result as a table or JSON based on the JSON flag
func FormatAuthResult(result AuthResult) error {
	if log.JSONFlag {
		// Use standard JSON encoding for single object output
		data := map[string]interface{}{
			"status":   result.Status,
			"node":     result.Node,
			"username": result.Username,
		}

		if result.Message != "" {
			data["message"] = result.Message
		}

		// Use standard JSON marshaling to output a single object
		jsonBytes, err := json.Marshal(data)
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %w", err)
		}

		fmt.Println(string(jsonBytes))
		return nil
	} else {
		// Use simple table format
		FormatAuthResultSimple(result)
		return nil
	}
}

// FormatError formats an error message as a table
func FormatError(operation, node, username, errorMsg string) error {
	result := AuthResult{
		Status:   "ERROR",
		Node:     node,
		Username: username,
		Message:  errorMsg,
	}
	return FormatAuthResult(result)
}

// FormatSuccess formats a success message as a table
func FormatSuccess(node, username string) error {
	result := AuthResult{
		Status:   "SUCCESS",
		Node:     node,
		Username: username,
		Message:  "Authentication successful",
	}
	return FormatAuthResult(result)
}

// FormatSimpleTable is a fallback function that prints a simple table format
func FormatSimpleTable(headers []string, rows [][]string) {
	// Print headers
	fmt.Printf("| ")
	for _, header := range headers {
		fmt.Printf("%-15s | ", header)
	}
	fmt.Println()

	// Print separator
	fmt.Printf("|")
	for range headers {
		fmt.Printf("-----------------|")
	}
	fmt.Println()

	// Print rows
	for _, row := range rows {
		fmt.Printf("| ")
		for _, cell := range row {
			fmt.Printf("%-15s | ", cell)
		}
		fmt.Println()
	}
}

// FormatAuthResultSimple formats an authentication result using simple table format as fallback
func FormatAuthResultSimple(result AuthResult) {
	headers := []string{"Status", "Node", "Username"}
	row := []string{result.Status, result.Node, result.Username}

	if result.Message != "" {
		headers = append(headers, "Message")
		row = append(row, result.Message)
	}

	FormatSimpleTable(headers, [][]string{row})
}
