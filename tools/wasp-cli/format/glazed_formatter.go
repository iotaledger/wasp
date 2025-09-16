package format

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-go-golems/glazed/pkg/middlewares"
	"github.com/go-go-golems/glazed/pkg/types"
	"github.com/iotaledger/wasp/v2/tools/wasp-cli/log"
)

// GlazedFormatter provides unified output formatting using glazed
type GlazedFormatter struct {
	processor *middlewares.TableProcessor
	jsonMode  bool
}

// NewGlazedFormatter creates a new glazed formatter
func NewGlazedFormatter() *GlazedFormatter {
	return &GlazedFormatter{
		processor: middlewares.NewTableProcessor(),
		jsonMode:  log.JSONFlag,
	}
}

// CommandData represents the standard structure for all command outputs
type CommandData struct {
	Type      string                 `json:"type"`
	Status    string                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// NewCommandData creates a new command data structure
func NewCommandData(commandType, status string, data map[string]interface{}) *CommandData {
	return &CommandData{
		Type:      commandType,
		Status:    status,
		Timestamp: time.Now().UTC(),
		Data:      data,
	}
}

// Format formats and outputs the command data
func (gf *GlazedFormatter) Format(commandData *CommandData) error {
	if gf.jsonMode {
		return gf.formatJSON(commandData)
	}
	return gf.formatTable(commandData)
}

// formatJSON outputs the data as JSON
func (gf *GlazedFormatter) formatJSON(commandData *CommandData) error {
	jsonBytes, err := json.MarshalIndent(commandData, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	fmt.Println(string(jsonBytes))
	return nil
}

// formatTable outputs the data as a table using glazed
func (gf *GlazedFormatter) formatTable(commandData *CommandData) error {
	ctx := context.Background()

	// Create a new processor for this output
	processor := middlewares.NewTableProcessor()

	// Add metadata row
	metadataRow := types.NewRow(
		types.MRP("type", commandData.Type),
		types.MRP("status", commandData.Status),
		types.MRP("timestamp", commandData.Timestamp.Format(time.RFC3339)),
	)

	if err := processor.AddRow(ctx, metadataRow); err != nil {
		return fmt.Errorf("failed to add metadata row: %w", err)
	}

	// Add data rows
	for key, value := range commandData.Data {
		dataRow := types.NewRow(
			types.MRP("field", key),
			types.MRP("value", fmt.Sprintf("%v", value)),
		)

		if err := processor.AddRow(ctx, dataRow); err != nil {
			return fmt.Errorf("failed to add data row: %w", err)
		}
	}

	// Render the table
	if err := processor.Close(ctx); err != nil {
		return fmt.Errorf("failed to render table: %w", err)
	}

	return nil
}

// FormatSuccess creates and formats a successful command output
func (gf *GlazedFormatter) FormatSuccess(commandType string, data map[string]interface{}) error {
	commandData := NewCommandData(commandType, "success", data)
	return gf.Format(commandData)
}

// FormatError creates and formats an error command output
func (gf *GlazedFormatter) FormatError(commandType string, errorMsg string) error {
	data := map[string]interface{}{
		"error": errorMsg,
	}
	commandData := NewCommandData(commandType, "error", data)
	return gf.Format(commandData)
}

// FormatAuthResult formats authentication results using glazed
func (gf *GlazedFormatter) FormatAuthResult(status, node, username, message string) error {
	data := map[string]interface{}{
		"node":     node,
		"username": username,
	}

	if message != "" {
		data["message"] = message
	}

	commandData := NewCommandData("auth", status, data)
	return gf.Format(commandData)
}

// FormatWalletBalance formats wallet balance results using glazed
func (gf *GlazedFormatter) FormatWalletBalance(addressIndex uint32, address string, balances interface{}) error {
	data := map[string]interface{}{
		"address_index": addressIndex,
		"address":       address,
		"balances":      balances,
	}

	commandData := NewCommandData("wallet_balance", "success", data)
	return gf.Format(commandData)
}

// FormatWalletAddress formats wallet address results using glazed
func (gf *GlazedFormatter) FormatWalletAddress(addressIndex uint32, address string) error {
	data := map[string]interface{}{
		"address_index": addressIndex,
		"address":       address,
	}

	commandData := NewCommandData("wallet_address", "success", data)
	return gf.Format(commandData)
}

// Global glazed formatter instance
var defaultGlazedFormatter = NewGlazedFormatter()

// FormatOutput formats output using the default glazed formatter
func FormatOutput(commandType, status string, data map[string]interface{}) error {
	commandData := NewCommandData(commandType, status, data)
	return defaultGlazedFormatter.Format(commandData)
}

// FormatSuccess formats a successful output using the default glazed formatter
func FormatSuccess(commandType string, data map[string]interface{}) error {
	return defaultGlazedFormatter.FormatSuccess(commandType, data)
}

// FormatError formats an error output using the default glazed formatter
func FormatError(commandType string, errorMsg string) error {
	return defaultGlazedFormatter.FormatError(commandType, errorMsg)
}
