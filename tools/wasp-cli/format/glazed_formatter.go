// Package format provides output formatting utilities for the wasp-cli tool.
//
// It contains helpers to render data in various human-friendly formats
// (such as tables and JSON), primarily used by command implementations
// across the CLI to present information consistently.
package format

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/iotaledger/wasp/v2/clients/iotagraphql"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/log"
)

// GlazedFormatter provides proper glazed-based output formatting
type GlazedFormatter struct {
	jsonMode    bool
	compactJSON bool
	tableMode   bool
}

// NewGlazedFormatter creates a new glazed formatter
func NewGlazedFormatter() *GlazedFormatter {
	return &GlazedFormatter{
		jsonMode:    log.JSONFlag || log.JSONCompactFlag,
		compactJSON: log.JSONCompactFlag,
		tableMode:   log.TableFlag,
	}
}

// NewGlazedFormatterWithOptions creates a new glazed formatter with specified options
func NewGlazedFormatterWithOptions(jsonMode, compactJSON, tableMode bool) *GlazedFormatter {
	return &GlazedFormatter{
		jsonMode:    jsonMode,
		compactJSON: compactJSON,
		tableMode:   tableMode,
	}
}

// FormatData formats data using proper glazed methods
func (gf *GlazedFormatter) FormatData(data map[string]interface{}) error {
	// Always check the current flag values for dynamic behavior
	if log.JSONFlag || log.JSONCompactFlag || gf.jsonMode {
		return gf.formatJSON(data)
	}
	if log.TableFlag || gf.tableMode {
		return gf.formatTable(data)
	}
	return gf.formatSimple(data)
}

// formatJSON outputs data as JSON
func (gf *GlazedFormatter) formatJSON(data map[string]interface{}) error {
	var jsonBytes []byte
	var err error

	// Check if compact JSON is requested (either via flag or formatter option)
	if log.JSONCompactFlag || gf.compactJSON {
		jsonBytes, err = json.Marshal(data)
	} else {
		// Default to pretty-printed JSON
		jsonBytes, err = json.MarshalIndent(data, "", "  ")
	}

	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(jsonBytes))
	return nil
}

// formatSimple outputs data as simple key-value pairs (default format)
func (gf *GlazedFormatter) formatSimple(data map[string]interface{}) error {
	if len(data) == 0 {
		return nil
	}

	// Filter out metadata fields for simple output
	filteredData := make(map[string]interface{})
	metadataFields := map[string]bool{
		"type":      true,
		"status":    true,
		"timestamp": true,
	}

	for k, v := range data {
		if !metadataFields[k] {
			filteredData[k] = v
		}
	}

	// Sort keys for consistent output
	keys := make([]string, 0, len(filteredData))
	for k := range filteredData {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Print key-value pairs
	for _, key := range keys {
		value := filteredData[key]
		fmt.Printf("%s: %v\n", key, value)
	}

	return nil
}

// wrapText wraps text to fit within a specified width
func wrapText(text string, width int) []string {
	if len(text) <= width {
		return []string{text}
	}

	var lines []string
	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{text}
	}

	currentLine := words[0]
	for _, word := range words[1:] {
		if len(currentLine)+1+len(word) <= width {
			currentLine += " " + word
		} else {
			lines = append(lines, currentLine)
			currentLine = word
		}
	}
	lines = append(lines, currentLine)

	return lines
}

// formatTable outputs data as a table using glazed with proper output handling
func (gf *GlazedFormatter) formatTable(data map[string]interface{}) error {
	ctx := context.Background()

	// Create a glazed TableProcessor
	processor := middlewares.NewTableProcessor()

	// Convert data to glazed row
	row := types.NewRowFromMap(data)

	// Add the row to the processor
	if err := processor.AddRow(ctx, row); err != nil {
		return fmt.Errorf("failed to add row: %w", err)
	}

	// Close the processor
	if err := processor.Close(ctx); err != nil {
		return fmt.Errorf("failed to close processor: %w", err)
	}

	if len(data) == 0 {
		return nil
	}

	// Compute columns and widths
	columns, maxWidths := computeColumnsAndWidths(data)

	// Render table
	renderTopBorder(columns, maxWidths)
	renderHeaders(columns, maxWidths)
	renderHeaderSeparator(columns, maxWidths)

	wrappedColumns, maxRows := computeWrappedColumns(data, columns, maxWidths)
	renderDataRows(columns, maxWidths, wrappedColumns, maxRows)
	renderBottomBorder(columns, maxWidths)

	return nil
}

// computeColumnsAndWidths returns sorted column names and their max widths with limits applied
func computeColumnsAndWidths(data map[string]interface{}) ([]string, map[string]int) {
	columns := make([]string, 0, len(data))
	maxWidths := make(map[string]int)

	for key := range data {
		columns = append(columns, key)
		headerWidth := len(key)
		value := fmt.Sprintf("%v", data[key])
		dataWidth := len(value)

		w := headerWidth
		if dataWidth > w {
			w = dataWidth
		}
		if w < 10 {
			w = 10
		}
		if w > 30 {
			w = 30
		}
		maxWidths[key] = w
	}

	sort.Strings(columns)
	return columns, maxWidths
}

func renderTopBorder(columns []string, maxWidths map[string]int) {
	fmt.Print("┌")
	for i, col := range columns {
		if i > 0 {
			fmt.Print("┬")
		}
		fmt.Print(strings.Repeat("─", maxWidths[col]+2))
	}
	fmt.Println("┐")
}

func renderHeaders(columns []string, maxWidths map[string]int) {
	fmt.Print("│")
	for i, colName := range columns {
		if i > 0 {
			fmt.Print("│")
		}
		fmt.Printf(" %-*s ", maxWidths[colName], colName)
	}
	fmt.Println("│")
}

func renderHeaderSeparator(columns []string, maxWidths map[string]int) {
	fmt.Print("├")
	for i, col := range columns {
		if i > 0 {
			fmt.Print("┼")
		}
		fmt.Print(strings.Repeat("─", maxWidths[col]+2))
	}
	fmt.Println("┤")
}

func computeWrappedColumns(data map[string]interface{}, columns []string, maxWidths map[string]int) (map[string][]string, int) {
	wrappedColumns := make(map[string][]string)
	maxRows := 1
	for _, colName := range columns {
		value := fmt.Sprintf("%v", data[colName])
		wrapped := wrapText(value, maxWidths[colName])
		wrappedColumns[colName] = wrapped
		if len(wrapped) > maxRows {
			maxRows = len(wrapped)
		}
	}
	return wrappedColumns, maxRows
}

func renderDataRows(columns []string, maxWidths map[string]int, wrappedColumns map[string][]string, maxRows int) {
	for row := 0; row < maxRows; row++ {
		fmt.Print("│")
		for i, colName := range columns {
			if i > 0 {
				fmt.Print("│")
			}
			var cellText string
			if row < len(wrappedColumns[colName]) {
				cellText = wrappedColumns[colName][row]
			}
			fmt.Printf(" %-*s ", maxWidths[colName], cellText)
		}
		fmt.Println("│")
	}
}

func renderBottomBorder(columns []string, maxWidths map[string]int) {
	fmt.Print("└")
	for i, col := range columns {
		if i > 0 {
			fmt.Print("┴")
		}
		fmt.Print(strings.Repeat("─", maxWidths[col]+2))
	}
	fmt.Println("┘")
}

// FormatSuccess formats successful command output
func (gf *GlazedFormatter) FormatSuccess(commandType string, data map[string]interface{}) error {
	outputData := map[string]interface{}{
		"type":      commandType,
		"status":    "success",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	// Merge command data
	for k, v := range data {
		outputData[k] = v
	}

	return gf.FormatData(outputData)
}

// FormatError formats error command output
func (gf *GlazedFormatter) FormatError(commandType string, errorMsg string) error {
	outputData := map[string]interface{}{
		"type":      commandType,
		"status":    "error",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"error":     errorMsg,
	}

	return gf.FormatData(outputData)
}

// FormatAuthResult formats authentication results
func (gf *GlazedFormatter) FormatAuthResult(status, node, username, message string) error {
	data := map[string]interface{}{
		"node":     node,
		"username": username,
	}

	if message != "" {
		data["message"] = message
	}

	if status == "success" {
		return gf.FormatSuccess("auth", data)
	}
	return gf.FormatError("auth", message)
}

// FormatWalletBalance formats wallet balance results
func (gf *GlazedFormatter) FormatWalletBalance(addressIndex uint32, address string, balances []*iotagraphql.Balance) error {
	data := map[string]interface{}{
		"address_index": addressIndex,
		"address":       address,
		"balances":      balances,
	}

	return gf.FormatSuccess("wallet_balance", data)
}

// FormatWalletAddress formats wallet address results
func (gf *GlazedFormatter) FormatWalletAddress(addressIndex uint32, address string) error {
	data := map[string]interface{}{
		"address_index": addressIndex,
		"address":       address,
	}

	return gf.FormatSuccess("wallet_address", data)
}

// Global formatter instance
var defaultFormatter = NewGlazedFormatter()

// FormatSuccess formats a successful command output with a given command type and data using the default formatter.
func FormatSuccess(commandType string, data map[string]interface{}) error {
	return defaultFormatter.FormatSuccess(commandType, data)
}

// FormatError formats an error message along with a command type for output handling. It uses the defaultFormatter instance.
func FormatError(commandType string, errorMsg string) error {
	return defaultFormatter.FormatError(commandType, errorMsg)
}

// FormatAuthResult formats the authentication result with status, node, username, and message for output.
func FormatAuthResult(status, node, username, message string) error {
	return defaultFormatter.FormatAuthResult(status, node, username, message)
}

// FormatWalletBalance formats and outputs wallet balance information for a specific address and index.
func FormatWalletBalance(addressIndex uint32, address string, balances []*iotagraphql.Balance) error {
	return defaultFormatter.FormatWalletBalance(addressIndex, address, balances)
}

// FormatWalletAddress formats and outputs a wallet address, given its index and address string, using the default formatter.
func FormatWalletAddress(addressIndex uint32, address string) error {
	return defaultFormatter.FormatWalletAddress(addressIndex, address)
}

// FormatAndExitWithError formats an error for glazed without exiting
func FormatAndExitWithError(cmd interface{}, err error) error {
	if err == nil {
		return nil
	}

	// Try to format the error using glazed
	if formatErr := FormatError("application", err.Error()); formatErr != nil {
		// If formatting fails, just print the error
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}

	return err
}
